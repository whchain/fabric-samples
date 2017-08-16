package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"time"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the wine structure, with 7 properties.  Structure tags are used by encoding/json library
type Wine struct {
	Owner        string `json:"owner"`
	Model        string `json:"model"`
	ProduceDate  string `json:"produce_date"`
	ProducePlace string `json:"produce_date"`
	OutDate      string `json:"out_date"`
	OutPlace     string `json:"out_place"`
	DeviceUid    string `json:"device_uid"`
}

type Device struct {
	Uid    string `json:"uid"`
	Model  string `json:"model"`
	Brand  string `json:"brand"`
	Status string `json:"status"`
}

type WineHistory struct {
	Timestamp string `json:"timestamp"`
	Wine      Wine `json:"wine"`
}

type WholeHistory struct {
	Device        Device `json:"device"`
	WineHistories []WineHistory `json:"wine_histories"`
}

/*
 * The Init method is called when the Smart Contract "chain" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "enrollDevice" {
		return s.enrollDevice(stub, args)
	} else if function == "enrollWine" {
		return s.enrollWine(stub, args)
	} else if function == "transferWine" {
		return s.transferWine(stub, args)
	} else if function == "queryAllCars" {
		return s.queryWine(stub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryWine(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	deviceAsBytes, _ := stub.GetState("device" + args[0])

	if deviceAsBytes == nil {
		return shim.Error("Device not enrolled")
	}

	device := Device{}
	json.Unmarshal(deviceAsBytes, device)

	if device.Status != "bind" {
		return shim.Error("Wine not enrolled")
	}

	resultsIterator, err := stub.GetHistoryForKey("wine" + args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	wholeHistory := WholeHistory{
		Device:        device,
		WineHistories: []WineHistory{},
	}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		wineAsBytes := response.Value
		wine := Wine{}
		json.Unmarshal(wineAsBytes, wine)

		wineHistory := WineHistory{
			Timestamp: time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String(),
			Wine:      wine,
		}

		wholeHistory.WineHistories = append(wholeHistory.WineHistories, wineHistory)

	}

	wholeHistoryAsBytes, _ := json.Marshal(wholeHistory)

	return shim.Success(wholeHistoryAsBytes)
}

func (s *SmartContract) transferWine(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	deviceAsBytes, _ := stub.GetState("device" + args[0])

	if deviceAsBytes == nil {
		return shim.Error("Device not enrolled")
	}

	device := Device{}
	json.Unmarshal(deviceAsBytes, device)

	if device.Status != "bind" {
		return shim.Error("Wine not enrolled")
	}

	wineAsBytes, _ := stub.GetState("wine" + args[0])
	wine := Wine{}
	json.Unmarshal(wineAsBytes, wine)

	wine.Owner = args[1]
	wineAsBytes, _ = json.Marshal(wine)
	stub.PutState("wine"+args[0], wineAsBytes)
	return shim.Success(nil)
}

func (s *SmartContract) enrollDevice(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	deviceAsBytes, _ := stub.GetState("device" + args[0])

	if deviceAsBytes != nil {
		return shim.Error("Device already enrolled")
	}

	var device = Device{args[0], args[1], args[2], "enrolled"}
	deviceAsBytes, _ = json.Marshal(device)
	stub.PutState("device"+args[0], deviceAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) enrollWine(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	deviceAsBytes, _ := stub.GetState("device" + args[0])

	if deviceAsBytes == nil {
		return shim.Error("Device not enrolled")
	}

	device := Device{}
	json.Unmarshal(deviceAsBytes, device)

	if device.Status != "enrolled" {
		return shim.Error("Device already used")
	}

	json.Unmarshal(deviceAsBytes, &device)
	device.Status = "bind"
	deviceAsBytes, _ = json.Marshal(device)
	stub.PutState("device"+args[0], deviceAsBytes)

	var wine = Wine{args[1], args[2], args[3], args[4], args[5], args[6], args[0]}
	wineAsBytes, _ := json.Marshal(wine)
	stub.PutState("wine"+args[0], wineAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
