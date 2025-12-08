package transform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oliveagle/jsonpath"
)

type DSLTransformer struct{}

func NewDSLTransformer() *DSLTransformer {
	return &DSLTransformer{}
}

func (t *DSLTransformer) Transform(data []byte, template map[string]interface{}) ([]byte, error) {
	return t.TransformWithContext(data, template, nil)
}

func (t *DSLTransformer) TransformWithContext(data []byte, template map[string]interface{}, contextData map[string]interface{}) ([]byte, error) {
	if len(template) == 0 {
		return data, nil
	}

	var sourceData interface{}
	if err := json.Unmarshal(data, &sourceData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	result, err := t.processTemplate(sourceData, template, contextData)
	if err != nil {
		return nil, err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return output, nil
}

func (t *DSLTransformer) processTemplate(sourceData interface{}, template map[string]interface{}, contextData map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range template {
		processed, err := t.processValue(sourceData, value, contextData)
		if err != nil {
			return nil, fmt.Errorf("failed to process key %s: %w", key, err)
		}
		result[key] = processed
	}

	return result, nil
}

func (t *DSLTransformer) processValue(sourceData interface{}, value interface{}, contextData map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return t.processString(sourceData, v, contextData)
	case map[string]interface{}:
		if jsonPath, ok := v["json.path"].(string); ok {
			return t.processArray(sourceData, jsonPath, v, contextData)
		}
		return t.processTemplate(sourceData, v, contextData)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			processed, err := t.processValue(sourceData, item, contextData)
			if err != nil {
				return nil, err
			}
			result[i] = processed
		}
		return result, nil
	default:
		return v, nil
	}
}

func (t *DSLTransformer) processString(sourceData interface{}, value string, contextData map[string]interface{}) (interface{}, error) {
	if strings.HasPrefix(value, "@ctx.") {
		if contextData == nil {
			return nil, nil
		}
		ctxPath := strings.TrimPrefix(value, "@ctx.")
		return t.getContextValue(contextData, ctxPath), nil
	}

	if !strings.HasPrefix(value, "$.") {
		return value, nil
	}

	if value == "$." {
		return sourceData, nil
	}

	result, err := jsonpath.JsonPathLookup(sourceData, value)
	if err != nil {
		return nil, nil
	}

	return result, nil
}

func (t *DSLTransformer) getContextValue(contextData map[string]interface{}, path string) interface{} {
	keys := strings.Split(path, ".")
	var current interface{} = contextData

	for _, key := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			return nil
		}
	}

	return current
}

func (t *DSLTransformer) processArray(sourceData interface{}, arrayPath string, itemTemplate map[string]interface{}, contextData map[string]interface{}) (interface{}, error) {
	arrayData, err := jsonpath.JsonPathLookup(sourceData, arrayPath)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup array path %s: %w", arrayPath, err)
	}

	arraySlice, ok := arrayData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("path %s does not point to an array", arrayPath)
	}

	result := make([]interface{}, 0, len(arraySlice))

	templateCopy := make(map[string]interface{})
	for k, v := range itemTemplate {
		if k != "json.path" {
			templateCopy[k] = v
		}
	}

	for _, item := range arraySlice {
		processedItem, err := t.processTemplate(item, templateCopy, contextData)
		if err != nil {
			return nil, fmt.Errorf("failed to process array item: %w", err)
		}
		result = append(result, processedItem)
	}

	return result, nil
}
