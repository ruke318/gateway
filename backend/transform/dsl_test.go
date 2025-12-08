package transform

import (
	"encoding/json"
	"testing"
)

func TestDSLTransformer_BasicMapping(t *testing.T) {
	transformer := NewDSLTransformer()

	sourceJSON := `{
		"name": "John",
		"email": "john@example.com",
		"age": 30
	}`

	template := map[string]interface{}{
		"username": "$.name",
		"email":    "$.email",
		"age":      "$.age",
	}

	result, err := transformer.Transform([]byte(sourceJSON), template)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if output["username"] != "John" {
		t.Errorf("Expected username to be 'John', got %v", output["username"])
	}
	if output["email"] != "john@example.com" {
		t.Errorf("Expected email to be 'john@example.com', got %v", output["email"])
	}
}

func TestDSLTransformer_FixedValues(t *testing.T) {
	transformer := NewDSLTransformer()

	sourceJSON := `{
		"code": 0,
		"message": "ok"
	}`

	template := map[string]interface{}{
		"code_success": "200",
		"code":         "$.code",
		"msg":          "$.message",
		"status":       "success",
	}

	result, err := transformer.Transform([]byte(sourceJSON), template)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if output["code_success"] != "200" {
		t.Errorf("Expected code_success to be '200', got %v", output["code_success"])
	}
	if output["status"] != "success" {
		t.Errorf("Expected status to be 'success', got %v", output["status"])
	}
}

func TestDSLTransformer_ArrayProcessing(t *testing.T) {
	transformer := NewDSLTransformer()

	sourceJSON := `{
		"code": 0,
		"message": "success",
		"data": [
			{
				"ID_SRV": "001",
				"EXAMINE_NAME": "Blood Test",
				"citem_type": "LAB",
				"SD_LABSAMP_CD": "BLD001",
				"HOSPITAL_DISTRICT_NO": "H001"
			},
			{
				"ID_SRV": "002",
				"EXAMINE_NAME": "X-Ray",
				"citem_type": "IMG",
				"SD_LABSAMP_CD": "XRY001",
				"HOSPITAL_DISTRICT_NO": "H002"
			}
		]
	}`

	template := map[string]interface{}{
		"code_success": "200",
		"code":         "$.code",
		"msg":          "$.message",
		"data": map[string]interface{}{
			"pages": "1",
			"zd_list": map[string]interface{}{
				"json.path":        "$.data",
				"page_no":          "1",
				"item_id":          "$.ID_SRV",
				"item_name":        "$.EXAMINE_NAME",
				"citem_type":       "$.citem_type",
				"lab_sample_code":  "$.SD_LABSAMP_CD",
				"hospital_no":      "$.HOSPITAL_DISTRICT_NO",
				"origin_data":      "$.",
			},
		},
	}

	result, err := transformer.Transform([]byte(sourceJSON), template)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if output["code_success"] != "200" {
		t.Errorf("Expected code_success to be '200', got %v", output["code_success"])
	}

	data, ok := output["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map")
	}

	if data["pages"] != "1" {
		t.Errorf("Expected pages to be '1', got %v", data["pages"])
	}

	zdList, ok := data["zd_list"].([]interface{})
	if !ok {
		t.Fatalf("Expected zd_list to be an array")
	}

	if len(zdList) != 2 {
		t.Errorf("Expected zd_list to have 2 items, got %d", len(zdList))
	}

	firstItem, ok := zdList[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected first item to be a map")
	}

	if firstItem["item_id"] != "001" {
		t.Errorf("Expected item_id to be '001', got %v", firstItem["item_id"])
	}
	if firstItem["item_name"] != "Blood Test" {
		t.Errorf("Expected item_name to be 'Blood Test', got %v", firstItem["item_name"])
	}
	if firstItem["page_no"] != "1" {
		t.Errorf("Expected page_no to be '1', got %v", firstItem["page_no"])
	}

	originData, ok := firstItem["origin_data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected origin_data to be a map")
	}
	if originData["ID_SRV"] != "001" {
		t.Errorf("Expected origin_data.ID_SRV to be '001', got %v", originData["ID_SRV"])
	}
}

func TestDSLTransformer_NestedObjects(t *testing.T) {
	transformer := NewDSLTransformer()

	sourceJSON := `{
		"orderId": "ORD123",
		"items": ["item1", "item2"],
		"customer": {
			"name": "Alice",
			"phone": "123-456-7890"
		},
		"createdAt": "2024-01-01"
	}`

	template := map[string]interface{}{
		"order": map[string]interface{}{
			"id":    "$.orderId",
			"items": "$.items",
			"customer": map[string]interface{}{
				"name":  "$.customer.name",
				"phone": "$.customer.phone",
			},
		},
		"timestamp": "$.createdAt",
	}

	result, err := transformer.Transform([]byte(sourceJSON), template)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	order, ok := output["order"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected order to be a map")
	}

	if order["id"] != "ORD123" {
		t.Errorf("Expected order.id to be 'ORD123', got %v", order["id"])
	}

	customer, ok := order["customer"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected order.customer to be a map")
	}

	if customer["name"] != "Alice" {
		t.Errorf("Expected customer.name to be 'Alice', got %v", customer["name"])
	}
}

func TestDSLTransformer_ContextDataAccess(t *testing.T) {
	transformer := NewDSLTransformer()

	sourceJSON := `{
		"code": 0,
		"message": "success",
		"data": {
			"id": 123,
			"name": "Test"
		}
	}`

	contextData := map[string]interface{}{
		"requestId": "req-12345",
		"tenantId":  "tenant-001",
		"user": map[string]interface{}{
			"id":   "user-999",
			"name": "John Doe",
		},
	}

	template := map[string]interface{}{
		"code":      "$.code",
		"message":   "$.message",
		"requestId": "@ctx.requestId",
		"tenantId":  "@ctx.tenantId",
		"userId":    "@ctx.user.id",
		"userName":  "@ctx.user.name",
		"dataId":    "$.data.id",
	}

	result, err := transformer.TransformWithContext([]byte(sourceJSON), template, contextData)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if output["requestId"] != "req-12345" {
		t.Errorf("Expected requestId to be 'req-12345', got %v", output["requestId"])
	}
	if output["tenantId"] != "tenant-001" {
		t.Errorf("Expected tenantId to be 'tenant-001', got %v", output["tenantId"])
	}
	if output["userId"] != "user-999" {
		t.Errorf("Expected userId to be 'user-999', got %v", output["userId"])
	}
	if output["userName"] != "John Doe" {
		t.Errorf("Expected userName to be 'John Doe', got %v", output["userName"])
	}
	if output["code"] != float64(0) {
		t.Errorf("Expected code to be 0, got %v", output["code"])
	}
	if output["dataId"] != float64(123) {
		t.Errorf("Expected dataId to be 123, got %v", output["dataId"])
	}
}

func TestDSLTransformer_ArrayWithContextData(t *testing.T) {
	transformer := NewDSLTransformer()

	sourceJSON := `{
		"code": 0,
		"message": "success",
		"data": [
			{
				"ID_SRV": "001",
				"EXAMINE_NAME": "Blood Test"
			},
			{
				"ID_SRV": "002",
				"EXAMINE_NAME": "X-Ray"
			}
		]
	}`

	contextData := map[string]interface{}{
		"requestId": "req-12345",
		"tenantId":  "tenant-001",
	}

	template := map[string]interface{}{
		"code_success": "200",
		"code":         "$.code",
		"msg":          "$.message",
		"requestId":    "@ctx.requestId",
		"data": map[string]interface{}{
			"pages": "1",
			"zd_list": map[string]interface{}{
				"json.path":   "$.data",
				"page_no":     "1",
				"item_id":     "$.ID_SRV",
				"item_name":   "$.EXAMINE_NAME",
				"tenant_id":   "@ctx.tenantId",
				"origin_data": "$.",
			},
		},
	}

	result, err := transformer.TransformWithContext([]byte(sourceJSON), template, contextData)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(result, &output); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if output["code_success"] != "200" {
		t.Errorf("Expected code_success to be '200', got %v", output["code_success"])
	}
	if output["requestId"] != "req-12345" {
		t.Errorf("Expected requestId to be 'req-12345', got %v", output["requestId"])
	}

	data, ok := output["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map")
	}

	zdList, ok := data["zd_list"].([]interface{})
	if !ok {
		t.Fatalf("Expected zd_list to be an array")
	}

	if len(zdList) != 2 {
		t.Errorf("Expected zd_list to have 2 items, got %d", len(zdList))
	}

	firstItem, ok := zdList[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected first item to be a map")
	}

	if firstItem["item_id"] != "001" {
		t.Errorf("Expected item_id to be '001', got %v", firstItem["item_id"])
	}
	if firstItem["tenant_id"] != "tenant-001" {
		t.Errorf("Expected tenant_id to be 'tenant-001', got %v", firstItem["tenant_id"])
	}
	if firstItem["page_no"] != "1" {
		t.Errorf("Expected page_no to be '1', got %v", firstItem["page_no"])
	}
}
