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
		fmt.Println("Getting State on fromCompany " + tr.FromCompany)	
		fromCompanyBytes, err := stub.GetState(accountPrefix+tr.FromCompany)
		if err != nil {
			fmt.Println("Account not found " + tr.FromCompany)
			return nil, errors.New("Account not found " + tr.FromCompany)
		}
		fmt.Println("---------------------transferPaper--------------part2---------success---")
		// Get account infromation of from company
		fmt.Println("Unmarshalling FromCompany ")
		err = json.Unmarshal(fromCompanyBytes, &fromCompany)
		if err != nil {
			fmt.Println("Error unmarshalling account " + tr.FromCompany)
			return nil, errors.New("Error unmarshalling account " + tr.FromCompany)
		}
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
		if err != nil {
			fmt.Println("Error unmarshalling account " + tr.ToCompany)
			return nil, errors.New("Error unmarshalling account " + tr.ToCompany)
		}

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
