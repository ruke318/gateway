package hook

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/model"
)

// DictionaryManager 字典管理器
// 负责加载、缓存和提供字典转换功能
type DictionaryManager struct {
	mu    sync.RWMutex
	cache map[string]map[string]map[string]string // orgID -> dictType -> key -> value
}

var globalDict = &DictionaryManager{
	cache: make(map[string]map[string]map[string]string),
}

// LoadDictionary 从数据库加载字典
// 在系统启动时调用，将所有字典配置加载到内存
func LoadDictionary() error {
	var configs []model.DictionaryConfig
	if err := database.DB.Find(&configs).Error; err != nil {
		return err
	}

	globalDict.mu.Lock()
	defer globalDict.mu.Unlock()

	// 清空旧缓存
	globalDict.cache = make(map[string]map[string]map[string]string)

	// 构建三层嵌套结构
	for _, cfg := range configs {
		if globalDict.cache[cfg.OrgID] == nil {
			globalDict.cache[cfg.OrgID] = make(map[string]map[string]string)
		}
		if globalDict.cache[cfg.OrgID][cfg.DictType] == nil {
			globalDict.cache[cfg.OrgID][cfg.DictType] = make(map[string]string)
		}
		globalDict.cache[cfg.OrgID][cfg.DictType][cfg.DictKey] = cfg.DictValue
	}

	return nil
}

// GetDictValue 获取字典值(机构内转换)
// 参数: orgID - 机构ID, dictType - 字典类型, key - 字典键
// 返回: 字典值，找不到则返回原 key
func (d *DictionaryManager) GetDictValue(orgID, dictType, key string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if orgDict, ok := d.cache[orgID]; ok {
		if typeDict, ok := orgDict[dictType]; ok {
			if val, ok := typeDict[key]; ok {
				return val
			}
		}
	}
	return key // 找不到则返回原值
}

// ReverseGetDictKey 反向查找：通过 value 反查 key
// 参数: orgID - 机构ID, dictType - 字典类型, value - 字典值
// 返回: 字典键，找不到则返回原 value
func (d *DictionaryManager) ReverseGetDictKey(orgID, dictType, value string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if orgDict, ok := d.cache[orgID]; ok {
		if typeDict, ok := orgDict[dictType]; ok {
			for k, v := range typeDict {
				if v == value {
					return k
				}
			}
		}
	}
	return value // 找不到则返回原值
}

// CrossOrgTranslate 跨机构字典转换
// 将 fromOrg 的值转换为 toOrg 的值
// 参数: fromOrg - 源机构ID, toOrg - 目标机构ID, dictType - 字典类型, value - 源机构的值
// 返回: 目标机构的值，找不到则返回原 value
func (d *DictionaryManager) CrossOrgTranslate(fromOrg, toOrg, dictType, value string) string {
	// 1. 反向查找 fromOrg 中的 key
	key := d.ReverseGetDictKey(fromOrg, dictType, value)
	if key == value {
		return value // 找不到 key，返回原值
	}

	// 2. 使用 key 在 toOrg 中查找值
	return d.GetDictValue(toOrg, dictType, key)
}

// GenerateDictionaryJS 生成 JS 字典函数
// 生成的函数会自动从 context 获取当前机构ID
func (d *DictionaryManager) GenerateDictionaryJS() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var script strings.Builder

	// 1. 序列化字典数据
	script.WriteString("var __dictData = ")
	jsonBytes, _ := json.Marshal(d.cache)
	script.Write(jsonBytes)
	script.WriteString(";\n\n")

	// 2. 定义字典对象和函数
	script.WriteString(`
// 字典对象
var dict = {};

// 获取当前机构ID（从 context.data.request.body.unit_id 获取）
dict.__getCurrentOrgID = function() {
    if (typeof context !== 'undefined' &&
        context.data &&
        context.data.request &&
        context.data.request.body &&
        context.data.request.body.unit_id) {
        return context.data.request.body.unit_id;
    }
    throw new Error("无法获取当前机构ID，请确保在 Hook 或 DSL 上下文中调用");
};

// 机构内字典转换（自动使用当前机构ID）
// 参数: dictType - 字典类型, key - 字典键
// 返回: 字典值（找不到返回原值）
// 示例: dict.get("payment_method", "ALIPAY") // 返回 "01"
dict.get = function(dictType, key) {
    var orgID = dict.__getCurrentOrgID();
    if (__dictData[orgID] && __dictData[orgID][dictType]) {
        return __dictData[orgID][dictType][key] || key;
    }
    return key;
};

// 反向查找：通过 value 反查 key
// 参数: dictType - 字典类型, value - 字典值
// 返回: 字典键（找不到返回原值）
// 示例: dict.reverseGet("payment_method", "01") // 返回 "ALIPAY"
dict.reverseGet = function(dictType, value) {
    var orgID = dict.__getCurrentOrgID();
    if (__dictData[orgID] && __dictData[orgID][dictType]) {
        for (var k in __dictData[orgID][dictType]) {
            if (__dictData[orgID][dictType][k] === value) {
                return k;
            }
        }
    }
    return value;
};

// 跨机构字典转换（fromOrg 默认使用当前机构）
// 参数: toOrg - 目标机构ID, dictType - 字典类型, value - 源机构的值
// 返回: 目标机构的值（找不到返回原值）
// 示例: dict.translate("org002", "order_status", "10") // 从当前机构转换到 org002
dict.translate = function(toOrg, dictType, value) {
    var fromOrg = dict.__getCurrentOrgID();

    // 1. 反向查找 fromOrg 的 key
    var key = null;
    if (__dictData[fromOrg] && __dictData[fromOrg][dictType]) {
        for (var k in __dictData[fromOrg][dictType]) {
            if (__dictData[fromOrg][dictType][k] === value) {
                key = k;
                break;
            }
        }
    }
    if (!key) return value;

    // 2. 使用 key 在 toOrg 中查找值
    if (__dictData[toOrg] && __dictData[toOrg][dictType]) {
        return __dictData[toOrg][dictType][key] || value;
    }
    return value;
};

// 完整跨机构转换（手动指定源机构和目标机构）
// 参数: fromOrg - 源机构ID, toOrg - 目标机构ID, dictType - 字典类型, value - 源机构的值
// 返回: 目标机构的值（找不到返回原值）
// 示例: dict.translateFull("org001", "org002", "order_status", "10")
dict.translateFull = function(fromOrg, toOrg, dictType, value) {
    // 1. 反向查找 fromOrg 的 key
    var key = null;
    if (__dictData[fromOrg] && __dictData[fromOrg][dictType]) {
        for (var k in __dictData[fromOrg][dictType]) {
            if (__dictData[fromOrg][dictType][k] === value) {
                key = k;
                break;
            }
        }
    }
    if (!key) return value;

    // 2. 使用 key 在 toOrg 中查找值
    if (__dictData[toOrg] && __dictData[toOrg][dictType]) {
        return __dictData[toOrg][dictType][key] || value;
    }
    return value;
};

// 批量转换（返回对象）
// 参数: dictType - 字典类型, keys - 键数组
// 返回: {key1: value1, key2: value2, ...}
// 示例: dict.batchGet("payment_method", ["ALIPAY", "WECHAT"])
dict.batchGet = function(dictType, keys) {
    var result = {};
    for (var i = 0; i < keys.length; i++) {
        var key = keys[i];
        result[key] = dict.get(dictType, key);
    }
    return result;
};

// 获取整个字典类型的所有映射
// 参数: dictType - 字典类型
// 返回: {key1: value1, key2: value2, ...}
// 示例: dict.getAll("payment_method") // 返回所有支付方式映射
dict.getAll = function(dictType) {
    var orgID = dict.__getCurrentOrgID();
    if (__dictData[orgID] && __dictData[orgID][dictType]) {
        return __dictData[orgID][dictType];
    }
    return {};
};
`)

	return script.String()
}

// ReloadDictionary 重新加载字典
// 在字典配置修改后调用，清空 VM 池以使新配置生效
func ReloadDictionary() error {
	if err := LoadDictionary(); err != nil {
		return err
	}
	// 清空 VM 池，让新请求使用新的字典
	if vmPool != nil {
		vmPool.Clear()
	}
	return nil
}

// GetDictionary 获取全局字典管理器
func GetDictionary() *DictionaryManager {
	return globalDict
}
