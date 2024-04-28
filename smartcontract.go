package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PaymentContract struct {
	contractapi.Contract
}
type SmartContract struct {
	contractapi.Contract
}

// details of the payment contract
type Contract struct {
	ID          string  `json:"ID"`          // Unique identifier for the contract
	Employer    string  `json:"Employer"`    // Name of the employer
	Employee    string  `json:"Employee"`    // Name of the employee
	Position    string  `json:"Position"`    // Position of the employee
	Salary      float64 `json:"Salary"`      // Annual salary of the employee
	VariablePay float64 `json:"VariablePay"` // Variable pay for the employee
	Currency    string  `json:"Currency"`    // Preferred currency for payment
	AccountID   string  `json:"Account"`     // Employee's bank account details
	Status      string  `json:"Status"`      // Status of the contract (active, revoked, etc.)
}

//details of a user account
type Account struct {
	AccountID         string `json:"AccountID"`         // Unique identifier for the account
	Company           string `json:"Company"`           // Company name
	TaxComplianceInfo string `json:"TaxComplianceInfo"` // Tax compliance information
	FinancialInfo     string `json:"FinancialInfo"`     // Confidential financial information
	PreferredCurrency string `json:"PreferredCurrency"` // Preferred currency for payment
	BankAccount       string `json:"BankAccount"`       // Bank account details
	ContractID        string `json:"ContractID"`        // ID of the associated contract
	ContractStatus    string `json:"ContractStatus"`    // Status of the associated contract
}

//details of an advance payment request
type AdvanceRequest struct {
	ID         string  `json:"ID"`
	ContractID string  `json:"ContractID"`
	Employee   string  `json:"Employee"`
	Amount     float64 `json:"Amount"`
	Status     string  `json:"Status"`
}

// payment transaction
type Payment struct {
	ID         string    `json:"ID"`
	ContractID string    `json:"ContractID"`
	Employee   string    `json:"Employee"`
	Amount     float64   `json:"Amount"`
	Date       time.Time `json:"Date"`
	Type       string    `json:"Type"`
}

// the interval for payroll payments
type PayrollInterval struct {
	StartDate time.Time `json:"StartDate"`
	EndDate   time.Time `json:"EndDate"`
}

// cross-border payment transaction
type CrossBorderPayment struct {
	ID         string  `json:"ID"`
	ContractID string  `json:"ContractID"`
	Employee   string  `json:"Employee"`
	Amount     float64 `json:"Amount"`
	Status     string  `json:"Status"`
}

// local payment transaction
type LocalPayment struct {
	ID         string  `json:"ID"`
	ContractID string  `json:"ContractID"`
	Employee   string  `json:"Employee"`
	Amount     float64 `json:"Amount"`
	Status     string  `json:"Status"`
}

// Constants for payment types
const (
	CrossBorder = "CrossBorder"
	Local       = "Local"
)

// Constants for payment types
const (
	RegularPayment = "Regular"
	AdvancePayment = "Advance"
)

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	contracts := []Asset{}

	return nil
}

// if a contract with the given ID exists
func (s *PaymentContract) ContractExists(ctx contractapi.TransactionContextInterface, contractID string) (bool, error) {
	contractJSON, err := ctx.GetStub().GetState(contractID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return contractJSON != nil, nil
}

// CreateContract creates a new payment contract between an employer and an employee
func (s *PaymentContract) CreateContract(ctx contractapi.TransactionContextInterface, contractID string, employer string, employee string, position string, salary float64, variablePay float64, currency string, account string ) error {
	exists, err := s.ContractExists(ctx, contractID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the contract %s already exists", contractID)
	}

	// Create new contract
	newContract := Contract{
		ID:          contractID,
		Employer:    employer,
		Employee:    employee,
		Position:    position,
		Salary:      salary,
		VariablePay: variablePay,
		Currency:    currency,
		Account:     account,
		Status:      "Active",
	}

	contractJSON, err := json.Marshal(newContract) //converting into JSON
	if err != nil {
		return err
	}

	// Put the contract on the ledger
	return ctx.GetStub().PutState(contractID, contractJSON)

}

// revoke an existing contract
func (s *PaymentContract) RevokeContract(ctx contractapi.TransactionContextInterface, contractID string) error {
	exists, err := s.ContractExists(ctx, contractID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the contract %s does not exist", contractID)
	}

	return ctx.GetStub().DelState(contractID)
}

//retrieves a contract by its ID
func (s *PaymentContract) GetContractByID(ctx contractapi.TransactionContextInterface, contractID string) (*Contract, error) {
	contractJSON, err := ctx.GetStub().GetState(contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if contractJSON == nil {
		return nil, fmt.Errorf("the contract %s does not exist", contractID)
	}

	var contract Contract
	err = json.Unmarshal(contractJSON, &contract)
	if err != nil {
		return nil, err
	}

	return &contract, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//Payroll
//////////////////////////////////////////////////////////////////////////////////////////////////

//monthly payment for an employee based on the contract details
func (s *PaymentContract) CalculateMonthlyPayment(contract *Contract) (float64, error) {
	monthlyPayment := contract.Salary + contract.VariablePay
	return monthlyPayment, nil
}

// new advance payment request
func (s *PaymentContract) AdvanceRequest(ctx contractapi.TransactionContextInterface, requestID string, contractID string, employee string, amount float64) error {
	// Check if contract exists
	contract, err := s.GetContractByID(ctx, contractID)
	if err != nil {
		return err
	}

	// monthly payment
	monthlyPayment, err := s.CalculateMonthlyPayment(contract)
	if err != nil {
		return err
	}

	// limits
	if amount > monthlyPayment*2 {
		return fmt.Errorf("advance amount exceeds limit")
	}

	// new advance request
	newRequest := AdvanceRequest{
		ID:         requestID,
		ContractID: contractID,
		Employee:   employee,
		Amount:     amount,
		Status:     "Pending", //yet to
	}

	requestJSON, err := json.Marshal(newRequest)
	if err != nil {
		return err
	}

	// Put the request on the ledger
	err = ctx.GetStub().PutState(requestID, requestJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// ApproveAdvanceRequest approves an advance payment request and processes the payment
func (s *PaymentContract) ApproveAdvanceRequest(ctx contractapi.TransactionContextInterface, requestID string) error {
	// Get advance request from the ledger
	requestJSON, err := ctx.GetStub().GetState(requestID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if requestJSON == nil {
		return fmt.Errorf("the advance request %s does not exist", requestID)
	}

	// Unmarshal advance request
	var request AdvanceRequest
	err = json.Unmarshal(requestJSON, &request)
	if err != nil {
		return err
	}

	// Update request status to Approved
	request.Status = "Approved"

	// Update request on the ledger
	requestJSON, err = json.Marshal(request)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(requestID, requestJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	// Process the advance payment
	err = s.ProcessPayment(ctx, request.ContractID, request.Employee, request.Amount, AdvancePayment)
	if err != nil {
		return err
	}

	return nil
}

// ProcessPayment processes a payment transaction
func (s *PaymentContract) ProcessPayment(ctx contractapi.TransactionContextInterface, contractID string, employee string, amount float64, paymentType string) error {
	// Check if contract exists
	contract, err := s.GetContractByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Calculate monthly payment for the contract
	monthlyPayment, err := s.CalculateMonthlyPayment(contract)
	if err != nil {
		return err
	}

	// Check if employee already received payment this month
	if paymentType == RegularPayment {
		lastPaymentDate, err := s.GetLastPaymentDate(ctx, contractID)
		if err != nil {
			return err
		}
		if lastPaymentDate.Month() == time.Now().Month() {
			return fmt.Errorf("employee already received payment this month")
		}
	}

	// Check if payment amount is within limits
	if amount > monthlyPayment*2 {
		return fmt.Errorf("payment amount exceeds limit")
	}

	// Create new payment transaction
	newPayment := Payment{
		ID:         fmt.Sprintf("PAY_%s_%s_%d", contractID, employee, time.Now().UnixNano()),
		ContractID: contractID,
		Employee:   employee,
		Amount:     amount,
		Date:       time.Now(),
		Type:       paymentType,
	}

	paymentJSON, err := json.Marshal(newPayment)
	if err != nil {
		return err
	}

	// Put the payment transaction on the ledger
	err = ctx.GetStub().PutState(newPayment.ID, paymentJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// WithdrawPayment withdraws the payment amount to the employee's designated account
func (s *PaymentContract) WithdrawPayment(ctx contractapi.TransactionContextInterface, contractID string, employee string, amount float64) error {
	// Check if contract exists contract,
	err := s.GetContractByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Get the latest payment transaction for the employee
	lastPayment, err := s.GetLastPayment(ctx, contractID, employee)
	if err != nil {
		return err
	}

	// Check if employee is trying to withdraw more than credited
	if amount > lastPayment.Amount {
		return fmt.Errorf("withdrawal amount exceeds credited amount")
	}

	// Create withdrawal transaction
	withdrawal := Payment{
		ID:         fmt.Sprintf("WITHDRAW_%s_%s_%d", contractID, employee, time.Now().UnixNano()),
		ContractID: contractID,
		Employee:   employee,
		Amount:     amount,
		Date:       time.Now(),
		Type:       "Withdrawal",
	}

	withdrawalJSON, err := json.Marshal(withdrawal)
	if err != nil {
		return err
	}

	// Put the withdrawal transaction on the ledger
	err = ctx.GetStub().PutState(withdrawal.ID, withdrawalJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// GetLastPaymentDate retrieves the last payment date for a contract
func (s *PaymentContract) GetLastPaymentDate(ctx contractapi.TransactionContextInterface, contractID string) (time.Time, error) {
	// Get all payment transactions for the contract
	paymentResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Payment", []string{contractID})
	if err != nil {
		return time.Time{}, err
	}
	defer paymentResultsIterator.Close()

	var lastPaymentDate time.Time
	for paymentResultsIterator.HasNext() {
		_, paymentKey, err := paymentResultsIterator.Next()
		if err != nil {
			return time.Time{}, err
		}

		// Extract the timestamp from the composite key
		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(paymentKey)
		if err != nil {
			return time.Time{}, err
		}

		// Convert the timestamp to time.Time
		timestamp, err := time.Parse(time.RFC3339Nano, compositeKeyParts[3])
		if err != nil {
			return time.Time{}, err
		}

		// Update lastPaymentDate if this payment is more recent
		if timestamp.After(lastPaymentDate) {
			lastPaymentDate = timestamp
		}
	}

	return lastPaymentDate, nil
}

// GetLastPayment retrieves the last payment transaction for an employee in a contract
func (s *PaymentContract) GetLastPayment(ctx contractapi.TransactionContextInterface, contractID string, employee string) (*Payment, error) {
	// Get all payment transactions for the contract and employee
	paymentResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Payment", []string{contractID, employee})
	if err != nil {
		return nil, err
	}
	defer paymentResultsIterator.Close()

	var lastPayment *Payment
	for paymentResultsIterator.HasNext() {
		_, paymentKey, err := paymentResultsIterator.Next()
		if err != nil {
			return nil, err
		}

		// Get the payment transaction
		paymentJSON, err := ctx.GetStub().GetState(paymentKey)
		if err != nil {
			return nil, err
		}

		var payment Payment
		err = json.Unmarshal(paymentJSON, &payment)
		if err != nil {
			return nil, err
		}

		// Update lastPayment if this payment is more recent
		if lastPayment == nil || payment.Date.After(lastPayment.Date) {
			lastPayment = &payment
		}
	}

	if lastPayment == nil {
		return nil, fmt.Errorf("no payments found for employee %s in contract %s", employee, contractID)
	}

	return lastPayment, nil
}

//################################################################################################
//################################################################################################
//Settlement
//################################################################################################
//################################################################################################

// ProcessPayment processes a payment transaction
func (s *PaymentContract) ProcessBankPayment(ctx contractapi.TransactionContextInterface, contractID string, employee string, amount float64, paymentType string) error {
	// Check if contract exists
	contract, err := s.GetContractByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Create new payment transaction
	var newPayment interface{}
	switch paymentType {
	case CrossBorder:
		newPayment = CrossBorderPayment{
			ID:         fmt.Sprintf("CROSS_%s_%s_%d", contractID, employee, time.Now().UnixNano()),
			ContractID: contractID,
			Employee:   employee,
			Amount:     amount,
			Status:     "Pending",
		}
	case Local:
		newPayment = LocalPayment{
			ID:         fmt.Sprintf("LOCAL_%s_%s_%d", contractID, employee, time.Now().UnixNano()),
			ContractID: contractID,
			Employee:   employee,
			Amount:     amount,
			Status:     "Pending",
		}
	default:
		return fmt.Errorf("invalid payment type")
	}

	paymentJSON, err := json.Marshal(newPayment)
	if err != nil {
		return err
	}

	// Put the payment transaction on the ledger
	err = ctx.GetStub().PutState(newPayment.ID, paymentJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// ApproveCrossBorderPayment approves a cross-border payment and processes the transaction
func (s *PaymentContract) ApproveCrossBorderPayment(ctx contractapi.TransactionContextInterface, paymentID string) error {
	// Get cross-border payment from the ledger
	paymentJSON, err := ctx.GetStub().GetState(paymentID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if paymentJSON == nil {
		return fmt.Errorf("the cross-border payment %s does not exist", paymentID)
	}

	// Unmarshal cross-border payment
	var payment CrossBorderPayment
	err = json.Unmarshal(paymentJSON, &payment)
	if err != nil {
		return err
	}

	// Approve the cross-border payment
	payment.Status = "Approved"

	// Update payment on the ledger
	paymentJSON, err = json.Marshal(payment)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(paymentID, paymentJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	// Process the cross-border payment (simulation)
	err = s.ProcessCrossBorderTransaction(ctx, payment)
	if err != nil {
		return err
	}

	return nil
}

// ProcessCrossBorderTransaction simulates the cross-border payment process
func (s *PaymentContract) ProcessCrossBorderTransaction(ctx contractapi.TransactionContextInterface, payment CrossBorderPayment) error {
	// In a real-world scenario, this function would interact with banks and forex services

	// Step 1: Central Bank "C" approves the transaction
	// Step 2: Central Bank "C" requests currency conversion from Forex Bank "B"
	// Step 3: Forex Bank "B" converts currency from currency "A" to "B"
	// Step 4: Central Bank of recipient nation receives converted amount
	// Step 5: Central Bank of recipient nation transfers amount to routing/member bank of payee

	// Simulating the process with logs
	fmt.Printf("Processing cross-border payment for contract %s, employee %s, amount %f\n", payment.ContractID, payment.Employee, payment.Amount)
	fmt.Println("Step 1: Central Bank C approves the transaction")
	fmt.Println("Step 2: Central Bank C requests currency conversion from Forex Bank B")
	fmt.Println("Step 3: Forex Bank B converts currency from currency A to B")
	fmt.Println("Step 4: Central Bank of recipient nation receives converted amount")
	fmt.Println("Step 5: Central Bank of recipient nation transfers amount to routing/member bank of payee")

	// Update payment status to completed
	payment.Status = "Completed"

	// Update payment on the ledger
	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(payment.ID, paymentJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// ProcessLocalPayment processes a local payment transaction
func (s *PaymentContract) ProcessLocalPayment(ctx contractapi.TransactionContextInterface, payment LocalPayment) error {
	// In a real-world scenario, this function would interact with local banks

	// Simulating the process with logs
	fmt.Printf("Processing local payment for contract %s, employee %s, amount %f\n", payment.ContractID, payment.Employee, payment.Amount)
	fmt.Println("Step 1: Bank C transfers money from party A to Bank D")
	fmt.Println("Step 2: Bank D credits amount to party B's account")

	// Update payment status to completed
	payment.Status = "Completed"

	// Update payment on the ledger
	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(payment.ID, paymentJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

/*
package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	AppraisedValue int    `json:"AppraisedValue"`
	Color          string `json:"Color"`
	ID             string `json:"ID"`
	Owner          string `json:"Owner"`
	Size           int    `json:"Size"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
		{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
		{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
		{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
		{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
		{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldOwner := asset.Owner
	asset.Owner = newOwner

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return oldOwner, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

*/
