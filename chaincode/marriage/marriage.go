package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Lover
}

type Lover struct {
	Name1   string `json:"name1"`
	Sex1    string `json:"sex1"`
	IDCard1 string `json:"idcard1"`

	Name2   string `json:"name2"`
	Sex2    string `json:"sex2"`
	IDCard2 string `json:"idcard2"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	lovers := []Lover{
		Lover{
			Name1:   "lilingj",
			Sex1:    "male",
			IDCard1: "123456xxxxxxxxxxxx",
			Name2:   "zhaox",
			Sex2:    "female",
			IDCard2: "654321xxxxxxxxxxxx",
		},
	}

	for _, lover := range lovers {
		loverAsBytes, _ := json.Marshal(lover)
		err := ctx.GetStub().PutState(lover.IDCard1+lover.IDCard2, loverAsBytes)

		if err != nil {
			return fmt.Errorf("初始化世界状态失败。 %s", err.Error())
		}
	}
	return nil
}

func (s *SmartContract) AddLover(ctx contractapi.TransactionContextInterface, name1 string, sex1 string, idcard1 string, name2 string, sex2 string, idcard2 string) error {
	lover := Lover{
		Name1:   name1,
		Sex1:    sex1,
		IDCard1: idcard1,
		Name2:   name2,
		Sex2:    sex2,
		IDCard2: idcard2,
	}

	loverAsBytes, _ := json.Marshal(lover)

	return ctx.GetStub().PutState(lover.IDCard1+lover.IDCard2, loverAsBytes)
}

func (s *SmartContract) QueryLover(ctx contractapi.TransactionContextInterface, idcard1 string, idcard2 string) (*Lover, error) {
	loverAsBytes, err := ctx.GetStub().GetState(idcard1 + idcard2)

	if err != nil {
		return nil, fmt.Errorf("Fled to read from world state. %s", err.Error())
	}

	if loverAsBytes == nil {
		return nil, fmt.Errorf("%s + %s does not exist", idcard1, idcard2)
	}

	lover := new(Lover)
	_ = json.Unmarshal(loverAsBytes, lover)

	return lover, nil
}

func (s *SmartContract) QueryAllLovers(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		lover := new(Lover)
		_ = json.Unmarshal(queryResponse.Value, lover)

		queryResult := QueryResult{Key: queryResponse.Key, Record: lover}
		results = append(results, queryResult)
	}
	return results, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
