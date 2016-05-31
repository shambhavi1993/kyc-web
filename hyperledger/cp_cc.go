	/*
	Copyright 2016 IBM

	Licensed under the Apache License, Version 2.0 (the "License")
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.

	Licensed Materials - Property of IBM
	© Copyright IBM Corp. 2016
	*/
	package main

	import (
		"encoding/json"
		"errors"
		"fmt"
		"strconv"
		"time"
		"strings"

		"github.com/hyperledger/fabric/core/chaincode/shim"
	)

	var cpPrefix = "cp:"
	var accountPrefix = "acct:"
	var accountsKey = "accounts"

	var recentLeapYear = 2016

	// SimpleChaincode example simple Chaincode implementation
	type SimpleChaincode struct {
	}

	func generateCUSIPSuffix(issueDate string, age string) (string, error) {

		t, err := msToTime(issueDate)
		if err != nil {
			return "", err
		}
		
		days, err := strconv.Atoi(age)
		if err != nil {
			// handle error
			fmt.Println(err)
			return "", err
		}

		maturityDate := t.AddDate(0, 0, days)
		month := int(maturityDate.Month())
		day := maturityDate.Day()

		suffix := seventhDigit[month] + eigthDigit[day]
		return suffix, nil

	}

	const (
		millisPerSecond     = int64(time.Second / time.Millisecond)
		nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
	)

	func msToTime(ms string) (time.Time, error) {
		msInt, err := strconv.ParseInt(ms, 10, 64)
		if err != nil {
			return time.Time{}, err
		}

		return time.Unix(msInt/millisPerSecond,
			(msInt%millisPerSecond)*nanosPerMillisecond), nil
	}

	type CP struct {
		CUSIP     string  `json:"cusip"`
		Name      string  `json:"ticker"`
		Gender    string  `json:"par"`
		Age       string  `json:"qty"`
		City  	  string  `json:"discount"`
		State	  string  `json:"maturity"`
		Owner     string  `json:"owner"`
		Issuer    string  `json:"issuer"`
		IssueDate string  `json:"issueDate"`
	}

	type Account struct {
		ID          string  `json:"id"`
		Prefix      string  `json:"prefix"`
		CashBalance float64 `json:"cashBalance"`
		AssetsIds   []string `json:"assetIds"`
	}

	type Transaction struct {
		CUSIP       string   `json:"cusip"`
		FromCompany string   `json:"fromCompany"`
		ToCompany   string   `json:"toCompany"`
		Quantity    int      `json:"quantity"`
		Discount    string   `json:"discount"`
	}

	func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
		// Initialize the collection of commercial paper keys
		fmt.Println("Initializing paper keys collection")
		var blank []string
		blankBytes, _ := json.Marshal(&blank)
		err := stub.PutState("PaperKeys", blankBytes)
		if err != nil {
			fmt.Println("Failed to initialize paper key collection")
		}

		fmt.Println("Initialization complete")
		return nil, nil
	}

	func (t *SimpleChaincode) createAccounts(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

		//  				0
		// "number of accounts to create"
		var err error
		numAccounts, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("error creating accounts with input")
			return nil, errors.New("createAccounts accepts a single integer argument")
		}
		//create a bunch of accounts
		var account Account
		counter := 1
		for counter <= numAccounts {
			var prefix string
			suffix := "000A"
			if counter < 10 {
				prefix = strconv.Itoa(counter) + "0" + suffix
			} else {
				prefix = strconv.Itoa(counter) + suffix
			}
			var assetIds []string
			account = Account{ID: "company" + strconv.Itoa(counter), Prefix: prefix, CashBalance: 10000000.0, AssetsIds: assetIds}
			accountBytes, err := json.Marshal(&account)
			if err != nil {
				fmt.Println("error creating account" + account.ID)
				return nil, errors.New("Error creating account " + account.ID)
			}
			err = stub.PutState(accountPrefix+account.ID, accountBytes)
			counter++
			fmt.Println("created account" + accountPrefix + account.ID)
		}

		fmt.Println("Accounts created") 
		return nil, nil

	}

	func (t *SimpleChaincode) createAccount(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
		// Obtain the username to associate with the account
		var account Account
	 if len(args) != 1 {
			fmt.Println("Error obtaining username")
			return nil, errors.New("createAccount accepts a single username argument")
		}
		username := args[0]
		
		// Build an account object for the user
		var assetIds []string
		suffix := "000A"
		prefix := username + suffix
		if username == "bank" {
					account = Account{ID: username, Prefix: prefix, CashBalance: 1000000.0, AssetsIds: assetIds}
			} else {
					account = Account{ID: username, Prefix: prefix, CashBalance: 100.0, AssetsIds: assetIds}
			}
		accountBytes, err := json.Marshal(&account)
		if err != nil {
			fmt.Println("error creating account" + account.ID)
			return nil, errors.New("Error creating account " + account.ID)
		}
		
		fmt.Println("Attempting to get state of any existing account for " + account.ID)
		existingBytes, err := stub.GetState(accountPrefix + account.ID)
		if err == nil {
			
			var company Account
			err = json.Unmarshal(existingBytes, &company)
			if err != nil {
				fmt.Println("Error unmarshalling account " + account.ID + "\n--->: " + err.Error())
				
				if strings.Contains(err.Error(), "unexpected end") {
					fmt.Println("No data means existing account found for " + account.ID + ", initializing account.")
					err = stub.PutState(accountPrefix+account.ID, accountBytes)
					
					if err == nil {
						fmt.Println("created account" + accountPrefix + account.ID)
						return nil, nil
					} else {
						fmt.Println("failed to create initialize account for " + account.ID)
						return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
					}
				} else {
					return nil, errors.New("Error unmarshalling existing account " + account.ID)
				}
			} else {
				fmt.Println("Account already exists for " + account.ID + " " + company.ID)
				return nil, errors.New("Can't reinitialize existing user " + account.ID)
			}
		} else {
			
			fmt.Println("No existing account found for " + account.ID + ", initializing account.")
			err = stub.PutState(accountPrefix+account.ID, accountBytes)
			
			if err == nil {
				fmt.Println("created account" + accountPrefix + account.ID)
				return nil, nil
			} else {
				fmt.Println("failed to create initialize account for " + account.ID)
				return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
			}
			
		} 

	}

	func (t *SimpleChaincode) issueCommercialPaper(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

		/*		0
			json
			{
				"ticker":  "string",
				"par": 0.00,
				"qty": 10,
				"discount": 7.5,
				"maturity": 30,
				"owners": [ // This one is not required
					{
						"company": "company1",
						"quantity": 5
					},
					{
						"company": "company3",
						"quantity": 3
					},
					{
						"company": "company4",
						"quantity": 2
					}
				],				
				"issuer":"company2",
				"issueDate":"1456161763790"  (current time in milliseconds as a string)

			}
		*/
		//need one arg
		if len(args) != 1 {
			fmt.Println("error invalid arguments")
			return nil, errors.New("Incorrect number of arguments. Expecting commercial paper record")
		}

		var cp CP
		var err error
		var account Account

		fmt.Println("Unmarshalling CP")
		err = json.Unmarshal([]byte(args[0]), &cp)
		if err != nil {
			fmt.Println("error invalid paper issue")
			return nil, errors.New("Invalid commercial paper issue")
		}
		//generate the CUSIP
		//get account prefix
		fmt.Println("Getting state of - " + accountPrefix + cp.Issuer)
		accountBytes, err := stub.GetState(accountPrefix + cp.Issuer)
		if err != nil {
			fmt.Println("Error Getting state of - " + accountPrefix + cp.Issuer)
			return nil, errors.New("Error retrieving account " + cp.Issuer)
		}
		err = json.Unmarshal(accountBytes, &account)
		if err != nil {
			fmt.Println("Error Unmarshalling accountBytes")
			return nil, errors.New("Error retrieving account " + cp.Issuer)
		}
		fmt.Println("-----------------Everything goes fine-------------")
		
		account.AssetsIds = append(account.AssetsIds, cp.CUSIP)

		// Set the issuer to be the owner of all quantity
		cp.Owner = "auditor_1"

		suffix, err := generateCUSIPSuffix(cp.IssueDate, cp.Age)
		if err != nil {
			fmt.Println("Error generating cusip")
			return nil, errors.New("Error generating CUSIP")
		}
		fmt.Println("Marshalling CP bytes")
		cp.CUSIP = account.Prefix + suffix
		fmt.Println("-----------------Everything goes fine-------------")
		fmt.Println("Getting State on CP " + cp.CUSIP)
		cpRxBytes, err := stub.GetState(cpPrefix+cp.CUSIP)
		if cpRxBytes == nil {
			fmt.Println("CUSIP does not exist, creating it")
			cpBytes, err := json.Marshal(&cp)
			if err != nil {
				fmt.Println("Error marshalling cp")
				return nil, errors.New("Error issuing commercial paper")
			}
			err = stub.PutState(cpPrefix+cp.CUSIP, cpBytes)
			if err != nil {
				fmt.Println("Error issuing paper")
				return nil, errors.New("Error issuing commercial paper")
			}

			fmt.Println("Marshalling account bytes to write")
			accountBytesToWrite, err := json.Marshal(&account)
			if err != nil {
				fmt.Println("Error marshalling account")
				return nil, errors.New("Error issuing commercial paper")
			}
			err = stub.PutState(accountPrefix + cp.Issuer, accountBytesToWrite)
			if err != nil {
				fmt.Println("Error putting state on accountBytesToWrite")
				return nil, errors.New("Error issuing commercial paper")
			}
			
			
			// Update the paper keys by adding the new key
			fmt.Println("Getting Paper Keys")
			keysBytes, err := stub.GetState("PaperKeys")
			if err != nil {
				fmt.Println("Error retrieving paper keys")
				return nil, errors.New("Error retrieving paper keys")
			}
			var keys []string
			err = json.Unmarshal(keysBytes, &keys)
			if err != nil {
				fmt.Println("Error unmarshel keys")
				return nil, errors.New("Error unmarshalling paper keys ")
			}
			
			fmt.Println("Appending the new key to Paper Keys")
			foundKey := false
			for _, key := range keys {
				if key == cpPrefix+cp.CUSIP {
					foundKey = true
				}
			}
			if foundKey == false {
				keys = append(keys, cpPrefix+cp.CUSIP)
				keysBytesToWrite, err := json.Marshal(&keys)
				if err != nil {
					fmt.Println("Error marshalling keys")
					return nil, errors.New("Error marshalling the keys")
				}
				fmt.Println("Put state on PaperKeys")
				err = stub.PutState("PaperKeys", keysBytesToWrite)
				if err != nil {
					fmt.Println("Error writting keys back")
					return nil, errors.New("Error writing the keys back")
				}
			}
			fmt.Println("--------------------------------------------------------Everything goes fine--------------------------------------------")
			fmt.Println("Issue commercial paper %+v\n", cp)
			return nil, nil
		} 
		return nil, nil
	}


	func GetAllCPs(stub *shim.ChaincodeStub) ([]CP, error){
	fmt.Println("--------------In GetAllCPs-------------")	
		var allCPs []CP
		
		// Get list of all the keys
		keysBytes, err := stub.GetState("PaperKeys")
		if err != nil {
			fmt.Println("Error retrieving paper keys")
			return nil, errors.New("Error retrieving paper keys")
		}
		var keys []string
		err = json.Unmarshal(keysBytes, &keys) 	
		if err != nil {
			fmt.Println("Error unmarshalling paper keys----------")
			fmt.Println(err)
			return nil, errors.New("Error unmarshalling paper keys")
		}

		// Get all the cps
		for _, value := range keys {
			fmt.Println("------------------------Keys-----------------"+value)
			cpBytes, err := stub.GetState(value)
			
			var cp CP
			err = json.Unmarshal(cpBytes, &cp)
			if err != nil {
				fmt.Println("Error retrieving cp " + value)
				return nil, errors.New("Error retrieving cp " + value)
			}
			
			fmt.Println("Appending CP" + value)
			allCPs = append(allCPs, cp)
		}	
		fmt.Println("-----------------------Everything goes fine in GetAllCPs------------------")
		return allCPs, nil 
	}

	func GetCP(cpid string, stub *shim.ChaincodeStub) (CP, error){
	fmt.Println("--------------In GetCP-------------")
		var cp CP
		cpBytes, err := stub.GetState(cpid)
		if err != nil {
			fmt.Println("Error retrieving cp " + cpid)
			return cp, errors.New("Error retrieving cp " + cpid)
		}
			
		err = json.Unmarshal(cpBytes, &cp)
		if err != nil {
			fmt.Println("Error unmarshalling cp " + cpid)
			return cp, errors.New("Error unmarshalling cp " + cpid)
		}
		return cp, nil 
	}


	func GetCompany(companyID string, stub *shim.ChaincodeStub) (Account, error){
	fmt.Println("--------------In GetCompany-------------")
		var company Account
		companyBytes, err := stub.GetState(accountPrefix+companyID)
		if err != nil {
			fmt.Println("Account not found " + companyID)
			return company, errors.New("Account not found " + companyID)
		}

		err = json.Unmarshal(companyBytes, &company)
		if err != nil {
			fmt.Println("Error unmarshalling account " + companyID + "\n err:" + err.Error())
			return company, errors.New("Error unmarshalling account " + companyID)
		}
		
		return company, nil 
	}


	// Still working on this one
	
	
	func (t *SimpleChaincode) transferPaper(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	fmt.Println("--------------In transferPaper-------------")
		/*		0
			json
			{
				  "CUSIP": "",
				  "fromCompany":"",
				  "toCompany":"",
				  "quantity": 1
			}
		*/
		//need one arg
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting commercial paper record")
		}
		fmt.Println("---------------------transferPaper--------------part0---------success---")
		var tr Transaction
		// Getting user input
		fmt.Println("Unmarshalling Transaction")
		err := json.Unmarshal([]byte(args[0]), &tr)
		if err != nil {
			fmt.Println("Error Unmarshalling Transaction")
			return nil, errors.New("Invalid commercial paper issue")
		}
		// Get state of CUSIP given by User
		fmt.Println("Getting State on CP " + tr.CUSIP)
		cpBytes, err := stub.GetState(cpPrefix+tr.CUSIP)
		if err != nil {
			fmt.Println("CUSIP not found")
			return nil, errors.New("CUSIP not found " + tr.CUSIP)
		}
		fmt.Println("---------------------transferPaper--------------part1---------success---")
		// Get Data of CUSIP from Blockchain
		var cp CP
		fmt.Println("Unmarshalling CP " + tr.CUSIP)
		err = json.Unmarshal(cpBytes, &cp)
		if err != nil {
			fmt.Println("Error unmarshalling cp " + tr.CUSIP)
			return nil, errors.New("Error unmarshalling cp " + tr.CUSIP)
		}
		// Get State for Account of from company
		var fromCompany Account
		fmt.Println("Getting State on fromCompany " + cp.Issuer)	
		fromCompanyBytes, err := stub.GetState(accountPrefix+cp.Issuer)
		if err != nil {
			fmt.Println("Account not found " + cp.Issuer)
			return nil, errors.New("Account not found " + cp.Issuer)
		}
		fmt.Println("---------------------transferPaper--------------part2---------success---")
		// Get account infromation of from company
		fmt.Println("Unmarshalling FromCompany ")
		err = json.Unmarshal(fromCompanyBytes, &fromCompany)
		if err != nil {
			fmt.Println("Error unmarshalling account " + tr.FromCompany)
			return nil, errors.New("Error unmarshalling account " + tr.FromCompany)
		}
		
			
		
	
		
		
		
		
		// Payment Transfer Start
		
	if tr.ToCompany	== "bank" {
			
		// Get state for Account of to company
		var toCompany Account
		fmt.Println("Getting State on ToCompany " + tr.ToCompany)
		toCompanyBytes, err := stub.GetState(accountPrefix+tr.ToCompany)
		if err != nil {
			fmt.Println("Account not found " + tr.ToCompany)
			return nil, errors.New("Account not found " + tr.ToCompany)
		}
		fmt.Println("---------------------transferPaper--------------part3---------success---")
		// Get Account infomation of to company
		fmt.Println("Unmarshalling tocompany")
		err = json.Unmarshal(toCompanyBytes, &toCompany)
		fmt.Println(err)
		fmt.Println(toCompanyBytes)
		fmt.Println("---------------------transferPaper--------------part4---------success---")
		if err != nil {
			fmt.Println("Error unmarshalling account " + tr.ToCompany)
			return nil, errors.New("Error unmarshalling account " + tr.ToCompany)
		}	
	
			amountToBeTransferred := 100.0
			
			// If toCompany doesn't have enough cash to buy the papers
			if toCompany.CashBalance < amountToBeTransferred {
				fmt.Println("The company " + tr.ToCompany + "doesn't have enough cash to purchase the papers")		
				return nil, errors.New("The company " + tr.ToCompany + "doesn't have enough cash to purchase the papers")	
			} else {
				fmt.Println("The ToCompany has enough money to be transferred for this paper")
			}
			
			toCompany.CashBalance -= amountToBeTransferred
			fromCompany.CashBalance += amountToBeTransferred

			toCompanyBytesToWrite, err := json.Marshal(&toCompany)
			fmt.Println("******************toCompanyData**********")
			fmt.Println(toCompanyBytesToWrite)
			fmt.Println("*****************************************")
			if err != nil {
				fmt.Println(err)
				fmt.Println("Error marshalling the toCompany")
				return nil, errors.New("Error marshalling the toCompany")
			}
			fmt.Println("Put state on toCompany")
			err = stub.PutState(accountPrefix+tr.ToCompany, toCompanyBytesToWrite)
			if err != nil {
				fmt.Println("Error writing the toCompany back")
				return nil, errors.New("Error writing the toCompany back")
			}
				
			// From company
			fromCompanyBytesToWrite, err := json.Marshal(&fromCompany)
			fmt.Println("******************toCompanyData**********")
			fmt.Println(fromCompany)
			fmt.Println("*****************************************")
			if err != nil {
				fmt.Println(err)
				fmt.Println("Error marshalling the fromCompany")
				return nil, errors.New("Error marshalling the fromCompany")
			}
			fmt.Println("Put state on fromCompany")
			err = stub.PutState(accountPrefix+cp.Issuer, fromCompanyBytesToWrite)
			if err != nil {
				fmt.Println("Error writing the fromCompany back")
				return nil, errors.New("Error writing the fromCompany back")
			}
	}		
		
		//Payment Transfer End
		
		
		
		
		
		
		
		
		
		

		// Check for all the possible errors
		ownerFound := false 
			if cp.Owner == tr.FromCompany {
				ownerFound = true
				cp.Owner = tr.ToCompany		//Transfer KYC to ToCompany
			}
		
		// If fromCompany doesn't own this paper
		if ownerFound == false {
			fmt.Println("The company " + tr.FromCompany + "doesn't own any of this paper")
			return nil, errors.New("The company " + tr.FromCompany + "doesn't own any of this paper")	
		} else {
			fmt.Println("The FromCompany does own this paper")
		}
		fmt.Println("---------------------transferPaper--------------part4---------success---")
		// cp
		cpBytesToWrite, err := json.Marshal(&cp)
		if err != nil {
			fmt.Println("Error marshalling the cp")
			return nil, errors.New("Error marshalling the cp")
		}
		fmt.Println("Put state on CP")
		err = stub.PutState(cpPrefix+tr.CUSIP, cpBytesToWrite)
		if err != nil {
			fmt.Println("Error writing the cp back")
			return nil, errors.New("Error writing the cp back")
		}
		
		fmt.Println("Successfully completed Invoke") 
		return nil, nil
	}


	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("----------------in Query------------")
		//need one arg
		if len(args) < 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting ......")
		}

		if args[0] == "GetAllCPs" {
			fmt.Println("Getting all CPs")
			allCPs, err := GetAllCPs(stub)
			if err != nil {
				fmt.Println("Error from getallcps")
				return nil, err
			} else {
				allCPsBytes, err1 := json.Marshal(&allCPs)
				if err1 != nil {
					fmt.Println("Error marshalling allcps")
					return nil, err1
				}	
				fmt.Println("All success, returning allcps")
				return allCPsBytes, nil		 
			}
		} else if args[0] == "GetCP" {
			fmt.Println("Getting particular cp")
			cp, err := GetCP(args[1], stub)
			if err != nil {
				fmt.Println("Error Getting particular cp")
				return nil, err
			} else {
				cpBytes, err1 := json.Marshal(&cp)
				if err1 != nil {
					fmt.Println("Error marshalling the cp")
					return nil, err1
				}	
				fmt.Println("All success, returning the cp")
				return cpBytes, nil		 
			}
		}	else if args[0] == "GetCompany" {
			fmt.Println("Getting the company")
			company, err := GetCompany(args[1], stub)
			if err != nil {
				fmt.Println("Error from getCompany")
				return nil, err
			} else {
				companyBytes, err1 := json.Marshal(&company)
				if err1 != nil {
					fmt.Println("Error marshalling the company")
					return nil, err1
				}	
				fmt.Println("All success, returning the company")
				return companyBytes, nil		 
			}
		} else {
			fmt.Println("Generic Query call")
			bytes, err := stub.GetState(args[0])

			if err != nil {
				fmt.Println("Some error happenend")
				return nil, errors.New("Some Error happened")
			}

			fmt.Println("All success, returning from generic")
			return bytes, nil		
		} 

		return nil, nil		//Added by ankit
	}

	func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
		fmt.Println("run is running " + function)
		return t.Invoke(stub, function, args)
	}

	func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
		fmt.Println("invoke is running " + function)
		
		if function == "issueCommercialPaper" {
			fmt.Println("Firing issueCommercialPaper")
			//Create an asset with some value
			return t.issueCommercialPaper(stub, args)
		} else if function == "transferPaper" {
			fmt.Println("Firing cretransferPaperateAccounts")
			return t.transferPaper(stub, args)
		} else if function == "createAccounts" {
			fmt.Println("Firing createAccounts")
			return t.createAccounts(stub, args)
		} else if function == "createAccount" {
			fmt.Println("Firing createAccount")
			return t.createAccount(stub, args)
		} else if function == "init" {
			fmt.Println("Firing init")
			return t.Init(stub, "init", args)
		}

		return nil, errors.New("Received unknown function invocation")
	}

	func main() {
		err := shim.Start(new(SimpleChaincode))
		if err != nil {
			fmt.Println("Error starting Simple chaincode: %s", err)
		}
	}

	//lookup tables for last two digits of CUSIP
	var seventhDigit = map[int]string{
		1:  "A",
		2:  "B",
		3:  "C",
		4:  "D",
		5:  "E",
		6:  "F",
		7:  "G",
		8:  "H",
		9:  "J",
		10: "K",
		11: "L",
		12: "M",
		13: "N",
		14: "P",
		15: "Q",
		16: "R",
		17: "S",
		18: "T",
		19: "U",
		20: "V",
		21: "W",
		22: "X",
		23: "Y",
		24: "Z",
	}

	var eigthDigit = map[int]string{
		1:  "1",
		2:  "2",
		3:  "3",
		4:  "4",
		5:  "5",
		6:  "6",
		7:  "7",
		8:  "8",
		9:  "9",
		10: "A",
		11: "B",
		12: "C",
		13: "D",
		14: "E",
		15: "F",
		16: "G",
		17: "H",
		18: "J",
		19: "K",
		20: "L",
		21: "M",
		22: "N",
		23: "P",
		24: "Q",
		25: "R",
		26: "S",
		27: "T",
		28: "U",
		29: "V",
		30: "W",
		31: "X",
	}
