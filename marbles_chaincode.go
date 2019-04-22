/*
 SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// 命令行的接口规范(接近web api的规范)：
// ==== Invoke orders ====
// 创建运单 peer chaincode invoke -C myc1 -n orders -c '{"Args":["initOrder","orderId0", "fromAddress", "toAddress", "煤炭", "20", "4000","WAIT_DRIVER_ACCEPT","goodsOwnerId","brokerId0","driverId"]}'
// 更改状态 peer chaincode invoke -C myc1 -n orders -c '{"Args":["changeStateOrder","orderId0","DRIVER_ACCEPT_WAIT_ROAD", "1553678282"]}'
// 更新位置 peer chaincode invoke -C myc1 -n orders -c '{"Args":["updatePositionOrder", "orderId0", "InterLocation1", "1553678182"]}'
// 删除运单 peer chaincode invoke -C myc1 -n orders -c '{"Args":["delete","orderId2"]}'

// ==== Query orders ====
// 从运单号查询运单 peer chaincode query -C myc1 -n orders -c '{"Args":["readOrder","orderId0"]}'
// 从运单号查询运单的轨迹 peer chaincode query -C myc1 -n orders -c '{"Args":["queryAssets","{\"selector\":{\"brokerId\":\"brokerId1\", \"docType\":\"position\"}}"]}'
// 查询该承运人时间戳范围内的运单 peer chaincode query -C myc1 -n orders -c '{"Args":["getOrdersByRange","",""]}'
// 从运单号查询运单的修改历史 peer chaincode query -C myc1 -n orders -c '{"Args":["getHistoryForOrder","order1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
// 从承运人查询运单列表 peer chaincode query -C myc1 -n orders -c '{"Args":["queryOrdersByBroker","brokerId1"]}'
// 从键值对查询运单列表 peer chaincode query -C myc1 -n orders -c '{"Args":["queryAssets","{\"selector\":{\"brokerId\":\"brokerId1\"}}"]}'

// Rich Query with Pagination (Only supported if CouchDB is used as state database):
// 查询分页的运单列表 peer chaincode query -C myc1 -n orders -c '{"Args":["queryOrdersWithPagination","{\"selector\":{\"docType\":\"order\"}}","3",""]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
// text file with extension *.json with the index definition formatted in JSON following the
// CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This orders02 example chaincode demonstrates a packaged
// index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
// to managed environments.
//
// In the examples below you can find index definitions that support orders02
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface or a curl command line utility.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, owner.
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"docType\",\"owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_orders/_index
//

// Index for docType, owner, size (descending order).
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"size\":\"desc\"},{\"docType\":\"desc\"},{\"owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_orders/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n orders -c '{"Args":["queryOrders","{\"selector\":{\"docType\":\"order\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n orders -c '{"Args":["queryOrders","{\"selector\":{\"docType\":{\"$eq\":\"order\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {	
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initOrder" { //create a new order
		return t.initOrder(stub, args)
	} else if function == "initStringHash" { //change owner of a specific order
		return t.initStringHash(stub, args)
	} else if function == "initFileHash" { //change owner of a specific order
		return t.initFileHash(stub, args)
	} else if function == "initUser" { //change owner of a specific order
		return t.initUser(stub, args)
	} else if function == "updateUser" { //change owner of a specific order
		return t.updateUser(stub, args)
	} else if function == "readUser" { //change owner of a specific order
		return t.readUser(stub, args)
	} else if function == "deleteUser" { //change owner of a specific order
		return t.deleteUser(stub, args)
	} else if function == "delete" { //delete a order
		return t.delete(stub, args)
	} else if function == "changeStateOrder" { //delete a order
		return t.changeStateOrder(stub, args)
	} else if function == "readOrder" { //read a order
		return t.readOrder(stub, args)
	} else if function == "queryOrdersByBroker" { //find orders for owner X using rich query
		return t.queryOrdersByBroker(stub, args)
	} else if function == "queryAssets" { //find orders based on an ad hoc rich query
		return t.queryAssets(stub, args)
	} else if function == "updatePositionOrder" { //find orders based on an ad hoc rich query
		return t.updatePositionOrder(stub, args)
	} else if function == "getHistoryForOrder" { //get history of values for a order
		return t.getHistoryForOrder(stub, args)
	} else if function == "getOrdersByRange" { //get orders based on range query
		return t.getOrdersByRange(stub, args)
	} else if function == "getOrdersByRangeWithPagination" {
		return t.getOrdersByRangeWithPagination(stub, args)
	} else if function == "queryOrderDetail" {
		return t.queryOrderDetail(stub, args)
	} else if function == "queryOrdersWithPagination" {
		return t.queryOrdersWithPagination(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func getTimeNow() string {
	var formatedTime string
	t := time.Now()
	formatedTime = t.Format(time.RFC1123)
	return formatedTime
}

// write to different ledgers- records, books and lending
func writeToRecordsLedger(stub shim.ChaincodeStubInterface, re Order, txnType string) pb.Response {
	if txnType != "createOrder" {
		//add TransactionHistory, first check if map has been initialized
		_, ok := re.ChangeStateHistory["createOrder"]
		if ok {
			re.ChangeStateHistory[txnType] = getTimeNow()
		} else {
			return shim.Error("......Records Transaction history is not initialized")
		}
	}
	// Encode JSON data
	reAsBytes, err := json.Marshal(re)

	// Store in the Blockchain
	err = stub.PutState(re.OrderId, reAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// ============================================================
// initOrder - create a new order, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//   0       		1       		2     	3		4      5 		6     					7     		8			9 
	// "orderId0", "fromAddress", "toAddress", "煤炭", "20", "4000","WAIT_DRIVER_ACCEPT","goodsOwnerId","brokerId0","driverId"
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}
	// ==== Input sanitation ====
	fmt.Println("- start init order")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	if len(args[7]) <= 0 {
		return shim.Error("8th argument must be a non-empty string")
	}
	if len(args[8]) <= 0 {
		return shim.Error("9th argument must be a non-empty string")
	}
	if len(args[9]) <= 0 {
		return shim.Error("10th argument must be a non-empty string")
	}
	orderId := args[0]
	fromAddress := args[1]
	toAddress := args[2]
	content := args[3]
	orderState := args[6]
	goodsOwnerId := args[7]
	brokerId := args[8]
	driverId := args[9]
	weightTon, err := strconv.ParseFloat(args[4], 64)
 	if err != nil {
		return shim.Error("5th argument must be a numeric string")
	}
	 transFee, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error("6th argument must be a numeric string")
	}

	// ==== Check if order already exists ====
	orderAsBytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Failed to get order: " + err.Error())
	} else if orderAsBytes != nil {
		fmt.Println("This order already exists: " + orderId)
		return shim.Error("This order already exists: " + orderId)
	}

	// ==== Create order object and marshal to JSON ====
	ChangeStateHistory := make(map[string]string)
	ChangeStateHistory["createOrder"] = getTimeNow()
	
	// order := &Order{"order","orderId0", "fromAddress", "toAddress", "coal", 20, 4000,"WAIT_DRIVER_ACCEPT","goodsOwnerId","brokerId0","driverId",getTimeNow(), true, ChangeStateHistory}

	order := &Order{"order", orderId, fromAddress, toAddress, content, weightTon, transFee, orderState, goodsOwnerId, brokerId, driverId, getTimeNow(), true, ChangeStateHistory}
	// writeToRecordsLedger(stub, order, "createOrder")
	orderJSONasBytes, err := json.Marshal(order)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the order json string manually if you don't want to use struct marshalling
	//orderJSONasString := `{"docType":"Order",  "name": "` + orderId + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//orderJSONasBytes := []byte(str)

	// === Save order to state ===
	err = stub.PutState(orderId, orderJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the order to enable color-based range queries, e.g. return all blue orders ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	// indexName := "broker~createDate"
	// dateBrokerIndexKey, err := stub.CreateCompositeKey(indexName, []string{order.BrokerId, order.CreateDate})
	// if err != nil {
	//	return shim.Error(err.Error())
	// }
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the order.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	// value := []byte{0x00}
	// stub.PutState(dateBrokerIndexKey, value)

	// ==== Order saved and indexed. Return success ====
	fmt.Println("- end init order")
	return shim.Success(nil)
}

func (t *SimpleChaincode) initStringHash(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       	1       2     			3			4  
	// "dataId", "orderId", "dataUrl", "shaResult", "comment"
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}
	// ==== Input sanitation ====
	fmt.Println("- start init order")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

	dataId := args[0]
	orderId := args[1]
	dataUrl := args[2]
	shaResult := args[3]
	comment := args[4]

	// ==== Check if order already exists ====
	orderAsBytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Failed to get order: " + err.Error())
	} else if orderAsBytes == nil {
		fmt.Println("This order not exists: " + orderId)
		return shim.Error("This order not exists: " + orderId)
	}

	// ==== Check if order already exists ====
	stringAsBytes, err := stub.GetState(dataId)
	if err != nil {
		return shim.Error("Failed to get string: " + err.Error())
	} else if stringAsBytes != nil {
		fmt.Println("This string already exists: " + dataId)
		return shim.Error("This string already exists: " + dataId)
	}

	// ==== Create marble object and marshal to JSON ====
	ObjectType := "stringHash"
	stringHash := &StringHash{ObjectType, dataId, orderId, dataUrl, shaResult, comment}
	stringHashJSONasBytes, err := json.Marshal(stringHash)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save marble to state ===
	err = stub.PutState(dataId, stringHashJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Order saved and indexed. Return success ====
	fmt.Println("- end init stringHash")
	return shim.Success(nil)
}

func (t *SimpleChaincode) initFileHash(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       		1       		2     	3		4  
	// "fileId", "orderId", "dataUrl", "shaResult", "comment"
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}
	// ==== Input sanitation ====
	fmt.Println("- start init order")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}

	fileId := args[0]
	orderId := args[1]
	dataUrl := args[2]
	shaResult := args[3]
	comment := args[4]

	isOrder, err := strconv.ParseBool(args[5])
	if err == nil {
	   /** displaying the string variable into the console */
	   fmt.Println("Value:", args[5])
	}

	// ==== Check if order already exists ====
	orderAsBytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Failed to get order: " + err.Error())
	} else if orderAsBytes == nil {
		fmt.Println("This order not exists: " + orderId)
		return shim.Error("This order not exists: " + orderId)
	}

	// ==== Check if order already exists ====
	fileAsBytes, err := stub.GetState(fileId)
	if err != nil {
		return shim.Error("Failed to get file: " + err.Error())
	} else if fileAsBytes != nil {
		fmt.Println("This file already exists: " + fileId)
		return shim.Error("This file already exists: " + fileId)
	}

	var ObjectType string
	// ==== Create marble object and marshal to JSON ====
	if isOrder == true {
		ObjectType = "fileHashForOrder"
	} else {
		ObjectType = "fileHashForUser"
	}
	fileHash := &FileHash{ObjectType, fileId, orderId, dataUrl, shaResult, comment}
	fileHashJSONasBytes, err := json.Marshal(fileHash)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save marble to state ===
	err = stub.PutState(fileId, fileHashJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Order saved and indexed. Return success ====
	fmt.Println("- end init stringHash")
	return shim.Success(nil)
}

func (t *SimpleChaincode) initUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       		1      	2     	3		4  
	// "userId", "userName", "role", "telephone", "valid"
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	// ==== Input sanitation ====
	fmt.Println("- start init order")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}

	userId := args[0]
	userName := args[1]
	role := args[2]
	telephone := args[3]

	valid, err := strconv.ParseBool(args[4])
	if err == nil {
	   /** displaying the string variable into the console */
	   fmt.Println("Value:", args[4])
	}

	// ==== Check if order already exists ====
	orderAsBytes, err := stub.GetState(userId)
	if err != nil {
		return shim.Error("Failed to get order: " + err.Error())
	} else if orderAsBytes != nil {
		fmt.Println("This user exists: " + userId)
		return shim.Error("This user exists: " + userId)
	}

	// ==== Check if order already exists ====
	userAsBytes, err := stub.GetState(userId)
	if err != nil {
		return shim.Error("Failed to get user: " + err.Error())
	} else if userAsBytes != nil {
		fmt.Println("This user already exists: " + userId)
		return shim.Error("This user already exists: " + userId)
	}

	// ==== Create marble object and marshal to JSON ====
	ObjectType := "user"
	user := &User{ObjectType, userId, userName, role, telephone, valid}
	userJSONasBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save marble to state ===
	err = stub.PutState(userId, userJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Order saved and indexed. Return success ====
	fmt.Println("- end init stringHash")
	return shim.Success(nil)
}

// ===========================================================
// change the state of a order by setting a new state on the order
// ===========================================================
func (t *SimpleChaincode) updateUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       		1      	2     	3		4  
	// "userId", "userName", "role", "telephone", "valid"
	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	userId := args[0]
	newName := args[1]
	newRole := args[2]
	newTelephone := args[3]

	newValid, err := strconv.ParseBool(args[4])
	if err == nil {
	   /** displaying the string variable into the console */
	   fmt.Println("Value:", args[4])
	}

	userAsBytes, err := stub.GetState(userId)
	if err != nil {
		return shim.Error("Failed to get order:" + err.Error())
	} else if userAsBytes == nil {
		return shim.Error("Order does not exist")
	}
	userToChangeState := User{}
	err = json.Unmarshal(userAsBytes, &userToChangeState) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	
	userToChangeState.UserName = newName
	userToChangeState.Role = newRole
	userToChangeState.Telephone = newTelephone
	userToChangeState.Valid = newValid

	// Encode JSON data
	userInputBytes, err := json.Marshal(userToChangeState)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Store in the Blockchain
	err = stub.PutState(userToChangeState.UserId, userInputBytes)
	
	fmt.Println("- end transferOrder (success)")
	return shim.Success(nil)
}

// ===============================================
// readOrder - read a order from chaincode state
// ===============================================
func (t *SimpleChaincode) readUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var userId string
	var err error
	userId = args[0]

	mainStruct := UserGenerated{StatusMessage: "Success"}
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the order to query")
	}

	queryFile := fmt.Sprintf("{\"selector\":{\"docType\":\"fileHashForUser\",\"orderId\":\"%s\"}}", userId)
	fileResults, err := getQueryResultForQueryString(stub, queryFile)
	if err != nil {
		return shim.Error(err.Error())
	}

	var fileWithKeys []FileWithKey
	err = json.Unmarshal(fileResults, &fileWithKeys)
	if err != nil {
		return shim.Error(err.Error())
	}

	for _,fileWithKey := range fileWithKeys {
		mainStruct.File = append(mainStruct.File, FileHash{ObjectType:fileWithKey.Record.ObjectType,
			FileId:fileWithKey.Record.FileId, OrderId:fileWithKey.Record.OrderId,
			DataUrl:fileWithKey.Record.DataUrl, ShaResult:fileWithKey.Record.ShaResult,
			Comment:fileWithKey.Record.Comment})
	}

	userAsBytes, err := stub.GetState(userId)
	if err != nil {
		return shim.Error("Failed to get order:" + err.Error())
	}
	user := User{}
	err = json.Unmarshal(userAsBytes, &user) //unmarshal it aka JSON.parse()
	mainStruct.User = user
	js, err := json.MarshalIndent(mainStruct, "", "  ")
	if err != nil {
		return shim.Error(err.Error())
	}
	
	return shim.Success([]byte(js))
}

// ===============================================
// readOrder - read a order from chaincode state
// ===============================================
func (t *SimpleChaincode) readOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var orderId, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the order to query")
	}

	orderId = args[0]
	valAsbytes, err := stub.GetState(orderId) //get the order from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + orderId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Order does not exist: " + orderId + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a order key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var orderJSON Order
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	orderId := args[0]
	// to maintain the color~name index, we need to read the order first and get its color
	valAsbytes, err := stub.GetState(orderId) //get the order from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + orderId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Order does not exist: " + orderId + "\"}"
		return shim.Error(jsonResp)
	}
	err = json.Unmarshal([]byte(valAsbytes), &orderJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + orderId + "\"}"
		return shim.Error(jsonResp)
	}
	err = stub.DelState(orderId) //remove the order from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "broker~createDate"
	dateBrokerIndexKey, err := stub.CreateCompositeKey(indexName, []string{orderJSON.BrokerId, orderJSON.CreateDate})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(dateBrokerIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ==================================================
// delete - remove a order key/value pair from state
// ==================================================
func (t *SimpleChaincode) deleteUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var userJSON User
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	userId := args[0]
	// to maintain the color~name index, we need to read the order first and get its color
	valAsbytes, err := stub.GetState(userId) //get the order from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + userId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Order does not exist: " + userId + "\"}"
		return shim.Error(jsonResp)
	}
	err = json.Unmarshal([]byte(valAsbytes), &userJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + userId + "\"}"
		return shim.Error(jsonResp)
	}
	err = stub.DelState(userId) //remove the order from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	return shim.Success(nil)
}

// ===========================================================
// change the state of a order by setting a new state on the order
// ===========================================================
func (t *SimpleChaincode) changeStateOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       		1
	// "orderId0", "DRIVER_ACCEPT_WAIT_ROAD"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	orderId := args[0]
	newState := strings.ToLower(args[1])
	fmt.Println("- start changeStateOrder ", orderId, newState)
	orderAsBytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Failed to get order:" + err.Error())
	} else if orderAsBytes == nil {
		return shim.Error("Order does not exist")
	}
	orderToChangeState := Order{}
	err = json.Unmarshal(orderAsBytes, &orderToChangeState) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	if newState == "SIGNED" {
		orderToChangeState.Open = false 
	}

	orderToChangeState.OrderState = newState //change the state
	writeToRecordsLedger(stub, orderToChangeState, newState)
	fmt.Println("- end transferOrder (success)")
	return shim.Success(nil)
}

func (t *SimpleChaincode) updatePositionOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	//   0       		1			2			3						4	
	// "positionId0", "orderId0", "2", "Mon, 02 Jan 2006 15:04:05 MST", "上海"
	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	fmt.Println("- start update position")

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}

	orderId := args[0]
	// ==== Check if order already exists ====
	orderAsBytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Failed to get order: " + err.Error())
	} else if orderAsBytes == nil {
		fmt.Println("This order does not exists: " + orderId)
		return shim.Error("This order does not exists: " + orderId)
	}

	// ==== Create marble object and marshal to JSON ====
	ObjectType := "position"
	position := &UpdatePositionHistory{ObjectType, args[0], args[1], args[2], args[3], args[4]}
	positionJSONasBytes, err := json.Marshal(position)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)

	// === Save marble to state ===
	err = stub.PutState(args[0], positionJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init position")
	return shim.Success(nil)
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults adds QueryResponseMetadata, which contains pagination
// info, to the constructed query results
// ===========================================================================================
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {

	buffer.WriteString("[{\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}]")

	return buffer
}

// ===========================================================================================
// getOrdersByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getOrdersByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("- getOrdersByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// transferOrdersBasedOnBroker will transfer orders of a given color to a certain new owner.
// Uses a GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getOrdersByBroker(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0    
	// "broker1"
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	broker := args[0]
	fmt.Println("- start getOrdersByBroker ", broker)

	// Query the color~name index by color
	// This will execute a key range query on all keys starting with 'color'
	orderByBrokerResultsIterator, err := stub.GetStateByPartialCompositeKey("broker~createDate", []string{broker})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer orderByBrokerResultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(orderByBrokerResultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(buffer.Bytes())
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryOrdersByBroker queries for orders based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryOrdersByBroker(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	broker := args[0]

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"order\",\"brokerId\":\"%s\"}}", broker)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (t *SimpleChaincode) queryOrderDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "orderId0"
	mainStruct := AutoGenerated{StatusMessage: "Success"}
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	orderId := args[0]
	
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"stringHash\",\"orderId\":\"%s\"}}", orderId)
	stringResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("stringResults stringResults ", string(stringResults))

	//var stringHashs map[string][]json.RawMessage
	var stringWithKeys []StringWithKey
	err = json.Unmarshal(stringResults, &stringWithKeys)
	if err != nil {
		return shim.Error(err.Error())
	}

	for _,stringWithKey := range stringWithKeys {
		mainStruct.String = append(mainStruct.String, StringHash{ObjectType:stringWithKey.Record.ObjectType,
			DataId:stringWithKey.Record.DataId, OrderId:stringWithKey.Record.OrderId,
			DataUrl:stringWithKey.Record.DataUrl, ShaResult:stringWithKey.Record.ShaResult,
			Comment:stringWithKey.Record.Comment})
	}

	queryFile := fmt.Sprintf("{\"selector\":{\"docType\":\"fileHashForOrder\",\"orderId\":\"%s\"}}", orderId)
	fileResults, err := getQueryResultForQueryString(stub, queryFile)
	if err != nil {
		return shim.Error(err.Error())
	}

	var fileWithKeys []FileWithKey
	err = json.Unmarshal(fileResults, &fileWithKeys)
	if err != nil {
		return shim.Error(err.Error())
	}

	for _,fileWithKey := range fileWithKeys {
		mainStruct.File = append(mainStruct.File, FileHash{ObjectType:fileWithKey.Record.ObjectType,
			FileId:fileWithKey.Record.FileId, OrderId:fileWithKey.Record.OrderId,
			DataUrl:fileWithKey.Record.DataUrl, ShaResult:fileWithKey.Record.ShaResult,
			Comment:fileWithKey.Record.Comment})
	}

	queryPosition := fmt.Sprintf("{\"selector\":{\"docType\":\"position\",\"orderId\":\"%s\"}}", orderId)
	positionResults, err := getQueryResultForQueryString(stub, queryPosition)
	if err != nil {
		return shim.Error(err.Error())
	}

	var positions []UpdatePositionHistory
	err = json.Unmarshal(positionResults, &positions)
	if err != nil {
		return shim.Error(err.Error())
	}

	for _,position := range positions {
		mainStruct.Position = append(mainStruct.Position, position)
	}

	orderAsBytes, err := stub.GetState(orderId)
	if err != nil {
		return shim.Error("Failed to get order:" + err.Error())
	}
	order := Order{}
	err = json.Unmarshal(orderAsBytes, &order) //unmarshal it aka JSON.parse()
	mainStruct.Order = order
	js, err := json.MarshalIndent(mainStruct, "", "  ")
	if err != nil {
		return shim.Error(err.Error())
	}
	
	return shim.Success([]byte(js))
}

// ===== Example: Ad hoc rich query ========================================================
// queryAssts uses a query string to perform a query for assets.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryOrdersForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// ====== Pagination =========================================================================
// Pagination provides a method to retrieve records with a defined pagesize and
// start point (bookmark).  An empty string bookmark defines the first "page" of a query
// result.  Paginated queries return a bookmark that can be used in
// the next query to retrieve the next page of results.  Paginated queries extend
// rich queries and range queries to include a pagesize and bookmark.
//
// Two examples are provided in this example.  The first is getOrdersByRangeWithPagination
// which executes a paginated range query.
// The second example is a paginated query for rich ad-hoc queries.
// =========================================================================================

// ====== Example: Pagination with Range Query ===============================================
// getOrdersByRangeWithPagination performs a range query based on the start & end key,
// page size and a bookmark.

// The number of fetched records will be equal to or lesser than the page size.
// Paginated range queries are only valid for read only transactions.
// ===========================================================================================
func (t *SimpleChaincode) getOrdersByRangeWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	startKey := args[0]
	endKey := args[1]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[2], 10, 20)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[3]
	resultsIterator, responseMetadata, err := stub.GetStateByRangeWithPagination(startKey, endKey, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}
	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)
	fmt.Printf("- getOrdersByRange queryResult:\n%s\n", bufferWithPaginationInfo.String())
	return shim.Success(buffer.Bytes())
}

// ===== Example: Pagination with Ad hoc Rich Query ========================================================
// queryOrdersWithPagination uses a query string, page size and a bookmark to perform a query
// for orders. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryOrdersForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// =========================================================================================
func (t *SimpleChaincode) queryOrdersWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//   0
	// "queryString"
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	queryString := args[0]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[1], 10, 16)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[2]
	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())
	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) getHistoryForOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	orderId := args[0]

	fmt.Printf("- start getHistoryForOrder: %s\n", orderId)

	resultsIterator, err := stub.GetHistoryForKey(orderId)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the order
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON order)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForOrder returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
