/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	//"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type PersonTransactionList struct {
	Nino         string        `json:"nino"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	TransactionID string  `json:"transactionID"`
	Amount        float64 `json:"amount"`
	CoverPeriod   string  `json:"coverPeriod"`
	OwningSystem  string  `json:"owningSystem"`
	PaymentStatus string  `json:"paymentStatus"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("hello_world", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "createPaymentTransaction" {
		return t.createPaymentTransaction(stub, args)
	} else if function == "modifyPaymentTransactionOwningSystem" {
		return t.modifyPaymentTransactionOwningSystem(stub, args)
	}

	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// write - invoke function to write key/value pair
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Create payment transactions for a person
func (t *SimpleChaincode) createPaymentTransaction(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running createPaymentTransaction()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]

	personTransactionListNew := PersonTransactionList{}

	err = json.Unmarshal([]byte(value), &personTransactionListNew)

	if err != nil {
		return nil, err
	}

	for _, newTransaction := range personTransactionListNew.Transactions {

		fmt.Printf("New Transaction records is: %s", newTransaction)
	}

	personTransactionListExisting := PersonTransactionList{}

	argsRead := []string{args[0]}

	valAsbytes := []byte{}

	valAsbytes, err = t.read(stub, argsRead)

	if err != nil {
		fmt.Printf("In t.read error: %s", err)
		return nil, err
	}

	if valAsbytes != nil {

		err = json.Unmarshal([]byte(valAsbytes), &personTransactionListExisting)

	} else {

		personTransactionListExisting.Nino = key
	}

	fmt.Printf("Old Transaction list is: %s", personTransactionListExisting)

	for _, newTransaction := range personTransactionListNew.Transactions {

		fmt.Printf("New Transaction ID second loop is: %s", newTransaction.TransactionID)

		transactionFound := false

		for _, oldTransaction := range personTransactionListExisting.Transactions {

			fmt.Printf("Old Transaction ID second loop is: %s", oldTransaction.TransactionID)

			if newTransaction.TransactionID == oldTransaction.TransactionID {
				transactionFound = true
				break
			}
		}

		if !transactionFound {

			personTransactionListExisting.Transactions =
				append(personTransactionListExisting.Transactions, newTransaction)
		}
	}

	fmt.Printf("Transaction to be created list is: %s", personTransactionListExisting)

	if err != nil {
		fmt.Printf("In json.Unmarshal error: %s", err)
		return nil, err
	}

	personTransactionListTobeCreatedAsBytes, err := json.Marshal(personTransactionListExisting)

	if err != nil {
		fmt.Printf("In json.Marshal error: %s", err)
		return nil, err
	}

	fmt.Printf("Transaction to be created list as bytes: %s", personTransactionListTobeCreatedAsBytes)

	err = stub.PutState(key, personTransactionListTobeCreatedAsBytes)

	if err != nil {
		fmt.Printf("In putstate error:  %s", err)
		return nil, err
	}
	return nil, nil
}

// Modify payment transactions for a person
func (t *SimpleChaincode) modifyPaymentTransactionOwningSystem(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running modifyPaymentTransactionOwningSystem()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]

	modifiedTransaction := Transaction{}

	err = json.Unmarshal([]byte(value), &modifiedTransaction)

	if err != nil {
		return nil, err
	}

	fmt.Printf("Modified Transaction records is: %s", modifiedTransaction)

	personTransactionListExisting := PersonTransactionList{}

	argsRead := []string{args[0]}

	valAsbytes := []byte{}

	valAsbytes, err = t.read(stub, argsRead)

	if err != nil {
		fmt.Printf("In t.read error: %s", err)
		return nil, err
	}

	if valAsbytes != nil {

		err = json.Unmarshal([]byte(valAsbytes), &personTransactionListExisting)

	} else {

		return nil, errors.New("No Payment records found for NINO: " + key)
	}

	fmt.Printf("Existing Transaction list is: %s", personTransactionListExisting)

	transactionFound := false

	for _, existingTransaction := range personTransactionListExisting.Transactions {

		fmt.Printf("Existing Transaction ID in loop is: %s", existingTransaction.TransactionID)

		if modifiedTransaction.TransactionID == existingTransaction.TransactionID {
			
			refExTrans := &existingTransaction
			
			refExTrans.OwningSystem = modifiedTransaction.OwningSystem
			
			transactionFound = true
			break
		}
	}

	if !transactionFound {

		return nil, errors.New("No Payment record found for the ID: " + modifiedTransaction.TransactionID)
	}

	personTransactionListTobeCreatedAsBytes, err := json.Marshal(personTransactionListExisting)

	if err != nil {
		fmt.Printf("In json.Marshal error: %s", err)
		return nil, err
	}

	fmt.Printf("Transaction to be created list as bytes: %s", personTransactionListTobeCreatedAsBytes)

	err = stub.PutState(key, personTransactionListTobeCreatedAsBytes)

	if err != nil {
		fmt.Printf("In putstate error:  %s", err)
		return nil, err
	}
	return nil, nil
}
