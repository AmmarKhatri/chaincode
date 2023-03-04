package chaincode

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type Bond struct {
	ID     string `json:"Id"`
	Amount int    `json:"Amount"`
	Owner  string `json:"Owner"`
	Issue  int    `json:"Issue"`
	Expiry int    `json:"Expiry"`
}

type Transaction struct {
	ID     string `json:"Id"`
	Seller string `json:"Seller"`
	Buyer  string `json:"Buyer"`
	B_Id   string `json:"BondId"`
	Time   int    `json:"Time"`
	IsMint bool   `json:"IsMint"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	bonds := []Bond{
		{ID: "b_1", Amount: 100, Issue: int(time.Now().Unix()), Owner: "Tomoko", Expiry: int(time.Now().Unix() + 3600*24*30)},
		{ID: "b_2", Amount: 1000, Issue: int(time.Now().Unix()), Owner: "Brad", Expiry: int(time.Now().Unix() + 3600*24*30)},
		{ID: "b_3", Amount: 2000, Issue: int(time.Now().Unix()), Owner: "Jin Soo", Expiry: int(time.Now().Unix() + 3600*24*30)},
		{ID: "b_4", Amount: 3000, Issue: int(time.Now().Unix()), Owner: "Max", Expiry: int(time.Now().Unix() + 3600*24*30)},
		{ID: "b_5", Amount: 2000, Issue: int(time.Now().Unix()), Owner: "Adriana", Expiry: int(time.Now().Unix() + 3600*24*30)},
		{ID: "b_6", Amount: 1000, Issue: int(time.Now().Unix()), Owner: "Michel", Expiry: int(time.Now().Unix() + 3600*24*30)},
	}

	for _, bond := range bonds {
		bondJSON, err := json.Marshal(bond)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(bond.ID, bondJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	//setting latest bond number
	num, _ := json.Marshal(6)
	err := ctx.GetStub().PutState("B_num", num)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	//setting latest transaction number
	num1, _ := json.Marshal(0)
	err = ctx.GetStub().PutState("T_num", num1)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	return nil
}

// can now check if government is the one to invoke the contract or not
func isAdmin(stub shim.ChaincodeStubInterface, ownerCertPEM []byte) (bool, error) {
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return false, err
	}

	block, _ := pem.Decode(creatorBytes)
	if block == nil {
		return false, errors.New("failed to decode PEM block containing creator certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse creator certificate: %s", err)
	}

	ownerCert, err := x509.ParseCertificate(ownerCertPEM)
	if err != nil {
		return false, fmt.Errorf("failed to parse owner certificate: %s", err)
	}

	return cert.Equal(ownerCert), nil
}

func (s *SmartContract) mintBond(ctx contractapi.TransactionContextInterface, number int, amount int, owner string, expiry int) error {
	//if condition for government invokation of function only
	// ownerCertPEM, _ := ioutil.ReadFile("/path/to/owner/certificate.pem") // parse correct path to the admin certificate
	// is, err := isAdmin(ctx.GetStub(), ownerCertPEM)                      // pass the PEM-encoded certificate of the owner here
	// if err != nil {
	// 	return err
	// }
	// if !is {
	// 	return fmt.Errorf("is government: %v", is)
	// }
	//___________________________
	bondNumJson, err := ctx.GetStub().GetState("B_num")
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	var bondNum int
	err = json.Unmarshal(bondNumJson, &bondNum)
	if err != nil {
		return err
	}

	tranNumJson, err := ctx.GetStub().GetState("T_num")
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	var tranNum int
	err = json.Unmarshal(tranNumJson, &tranNum)
	if err != nil {
		return err
	}
	for i := bondNum + 1; i <= bondNum+number; i++ {
		id := fmt.Sprint("b_", i)
		//each bond reference
		bond := Bond{
			ID:     id,
			Amount: amount,
			Owner:  "Government", //change it to its address
			Issue:  int(time.Now().Unix()),
			Expiry: expiry,
		}
		//each transaction reference
		id1 := fmt.Sprint("transaction", i)
		transaction := Transaction{
			ID:     id1,
			Seller: "none",
			Buyer:  "none",
			B_Id:   id,
			Time:   int(time.Now().Unix()),
			IsMint: true,
		}
		//saving bond
		bondJSON, err := json.Marshal(bond)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(bond.ID, bondJSON)
		if err != nil {
			return fmt.Errorf("failed to put bond to world state. %v", err)
		}
		//saving transaction
		tranJSON, err := json.Marshal(transaction)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(transaction.ID, tranJSON)
		if err != nil {
			return fmt.Errorf("failed to put transaction to world state. %v", err)
		}
	}
	//sets the new bond number
	new := bondNum + number
	num, _ := json.Marshal(new)
	err = ctx.GetStub().PutState("B_num", num)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	// sets the new transaction number
	new1 := tranNum + number
	num1, _ := json.Marshal(new1)
	err = ctx.GetStub().PutState("T_num", num1)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	return nil
}

func (s *SmartContract) getBond(ctx contractapi.TransactionContextInterface, id string) (*Bond, error) {
	bondJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if bondJSON == nil {
		return nil, fmt.Errorf("the bond %s does not exist", id)
	}
	var bond Bond
	err = json.Unmarshal(bondJSON, &bond)
	if err != nil {
		return nil, err
	}

	return &bond, nil
}

func (s *SmartContract) getTransaction(ctx contractapi.TransactionContextInterface, id string) (*Transaction, error) {
	tranJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if tranJSON == nil {
		return nil, fmt.Errorf("the bond %s does not exist", id)
	}
	var tran Transaction
	err = json.Unmarshal(tranJSON, &tran)
	if err != nil {
		return nil, err
	}

	return &tran, nil
}

func (s *SmartContract) buyBond(ctx contractapi.TransactionContextInterface, id string, n_owner string) (string, error) {
	bond, err := s.getBond(ctx, id)
	if err != nil {
		return "nil", err
	}
	if bond.Expiry < int(time.Now().Unix()) {
		return "nil", fmt.Errorf("expired bond, cannot buy")
	}
	//getting transaction number
	tranNumJson, err := ctx.GetStub().GetState("T_num")
	if err != nil {
		return "nil", fmt.Errorf("failed to read from world state: %v", err)
	}
	var tranNum int
	err = json.Unmarshal(tranNumJson, &tranNum)
	if err != nil {
		return "nil", err
	}
	//made changes to bond
	seller := bond.Owner
	bond.Owner = n_owner
	newTransId := fmt.Sprintf("t_", tranNum+1)
	transaction := Transaction{
		ID:     newTransId,
		Seller: seller,
		Buyer:  n_owner,
		B_Id:   bond.ID,
		Time:   int(time.Now().Unix()),
		IsMint: false,
	}
	//saving transaction
	tranJSON, err := json.Marshal(transaction)
	if err != nil {
		return "nil", err
	}
	err = ctx.GetStub().PutState(transaction.ID, tranJSON)
	if err != nil {
		return "nil", fmt.Errorf("failed to put transaction to world state. %v", err)
	}
	// sets the new transaction number
	new1 := tranNum + 1
	num1, _ := json.Marshal(new1)
	err = ctx.GetStub().PutState("T_num", num1)
	if err != nil {
		return "nil", fmt.Errorf("failed to put to world state. %v", err)
	}
	return "Transaction added successfully", nil
}
