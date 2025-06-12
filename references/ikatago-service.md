# MGO Usages in ikatago-service 

---

### `/Users/jinggangwang/gochess/ikatago-service/credit/credit.go`

#### Unit Tests ####
should add unit test for index = mgo.Index{
		Key: []string{"userId", "-createdAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
make sure the created index name is "userId_1_createdAt_-1"

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_CREDITS_COLLECTION)

	index := mgo.Index{
		Key: []string{"userId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"userId", "creditType"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"creditType"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"connectUserId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"userId", "-createdAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	return nil
}
``` 

---

### `/Users/jinggangwang/gochess/ikatago-service/vip/service.go`

#### Unit Tests ####
1. should add unit test to test the sort, limit, one methods are working correctly
2. should add unit test to test the Find().Select().All() methods are working correctly. 
3. should add unit tests to test the pipe methods with similar parameters in the following code are working correctly.
4. should add unit tests to test the upsert methods are working correctly.
5. should add unit tests to test the count method is working correctly.

```go
func (service *Service) Init() error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)
	data := &VIPStatData{}
	err := c.Find(bson.M{}).Sort("-updatedAt").Limit(1).One(data)
	// ... existing code ...
}
```

```go
func (service *Service) checkVIPAutoRenew() {
	// ... existing code ...
	uc := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	now := time.Now()
	autoRenewUsers := make([]UserAccount, 0)
	err := uc.Find(bson.M{
		"membershipExpiresAt": bson.M{
			"$lte": now.Add(24 * time.Hour),
			"$gte": now.Add(24 * time.Hour * -1),
		},
		"membershipAutoRenew": true,
	}).Select(bson.M{"_id": 1}).All(&autoRenewUsers)
	// ... existing code ...
}
```

```go
func (service *Service) recalcuateLatestData() error {
	// ... existing code ...
	mc := session.DB("").C(utils.IKATAGO_CLUSTER_USER_MEMBERSHIP_RECORD_COLLECTION)
	records := make([]product.MembershipRecord, 0)
	err := mc.Find(bson.M{}).All(&records)
	// ... existing code ...
	uc := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	// ... existing code ...
	err = uc.Find(bson.M{
		"vip":      true,
		"finished": true,
		"serialId": bson.M{
			"$gt": currentSerialID,
		},
	}).Select(bson.M{"serialId": 1}).Sort("-serialId").Limit(1).One(&lastFinishedUsage)
	// ... existing code ...
	err = uc.Find(bson.M{
		"vip":      true,
		"finished": false,
		"serialId": bson.M{
			"$gt": currentSerialID,
			"$lt": maxSerialID,
		},
	}).Select(bson.M{"serialId": 1}).Sort("serialId").Limit(1).One(&minUnfinishedUsage)
	// ... existing code ...
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"serialId": bson.M{
					"$gt": currentSerialID,
					"$lt": maxSerialID,
				},
				"finished": true,
				"vip":      true,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"totalUsedComsumption": bson.M{
					"$sum": "$virtualTotalCost",
				},
				"maxSerialId": bson.M{
					"$max": "$serialId",
				},
			},
		},
	}
	resp := []bson.M{}
	err = uc.Pipe(pipeline).All(&resp)
	// ... existing code ...
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"serialId": bson.M{
					"$gt": newCheckpointedSerialID,
				},
				"vip": true,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"totalUsedComsumption": bson.M{
					"$sum": "$virtualTotalCost",
				},
			},
		},
	}
	resp2 := []bson.M{}
	err = uc.Pipe(pipeline2).All(&resp2)
	// ... existing code ...
	onlineVIPCount, err := uc.Find(bson.M{
		"vip":      true,
		"finished": false,
	}).Count()
	// ... existing code ...
	onlineJailedVIPCount, err := uc.Find(bson.M{
		"vip":      true,
		"finished": false,
		"jailed":   true,
	}).Count()
	// ... existing code ...
	vipc := session.DB("").C(utils.IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)
	_, err = vipc.Upsert(bson.M{
		"date": newData.Date,
	}, newData)
	// ... existing code ...
}
``` 

---

### `/Users/jinggangwang/gochess/ikatago-service/vip/vip.go`

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)

	index := mgo.Index{
		Key:    []string{"date"},
		Unique: true,
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"-updatedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	return nil
}
``` 

---

### `/Users/jinggangwang/gochess/ikatago-service/auth/account_service.go`
### Unit Tests ###
1. Add unit tests the update({_id: xxx}, {$set: {...}}) is working correctly.
2. Add unit tests to test: err := vc.Find(query).Sort("-createdAt").Limit(1).One(&code)
with reverse order sort and limit
3. add unit tests to test err = c.UpdateId(existingAccount.ID, existingAccount) updateId method
4. add pipe unit tests to test the Pipe function with the following pipe args:
	pipeline := []bson.M{
		{
			"$match": query,
		},
		{
			"$project": bson.M{
				"yearMonthDay": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "timezone": "+08:00", "date": "$createdAt"},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$yearMonthDay",
				"totalNewUsers": bson.M{
					"$sum": 1,
				},
			},
		},
	}
5. Add unit tests to test the removeId and removeAll

```go
func (accountService *AccountService) createActivateCode(activateCodeTTL int) (*AccountActivateCode, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACTIVATE_CODE_COLLECTION)
	// ... existing code ...
	err := c.Insert(activateCode)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) loginByActivateCode(activateCode string, mobileUUID string) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACTIVATE_CODE_COLLECTION)
	uc := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)

	activateCodeItem := AccountActivateCode{}
	err := c.Find(bson.M{"activateCode": activateCode}).One(&activateCodeItem)
	// ... existing code ...
		err := uc.Update(bson.M{"_id": account.ID}, bson.M{"$set": bson.M{
			"token":       account.Token,
			"updatedAt":   time.Now().Truncate(time.Millisecond),
			"lastLoginAt": time.Now().Truncate(time.Millisecond),
		},
		})
	// ... existing code ...
		err = c.Update(bson.M{"_id": activateCodeItem.ID}, activateCodeItem)
	// ... existing code ...
		err := uc.Update(bson.M{"_id": account.ID}, bson.M{"$set": bson.M{
			"lastLoginAt": time.Now().Truncate(time.Millisecond),
		},
		})
	// ... existing code ...
	err = uc.Insert(account)
	// ... existing code ...
	err = c.Update(bson.M{"_id": activateCodeItem.ID}, activateCodeItem)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) Login(method string, identifier string, password string) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	if method == "phone" {
		err = c.Find(bson.M{"phone": identifier}).One(&account)
	} else if method == "email" {
		err = c.Find(bson.M{"email": identifier}).One(&account)
	}
	// ... existing code ...
	_ = c.Update(bson.M{"_id": account.ID}, bson.M{"$set": bson.M{
		"lastLoginAt": account.LastLoginAt}})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) FastLogin(method string, identifier string, verificationCode string, checkVerificationCode bool) (*Account, error) {
	// ... existing code ...
	// ... existing code ...
		vc := session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
		// ... existing code ...
		query := bson.M{
			"code": verificationCode,
			"type": "fast_login",
			"expiresAt": bson.M{
				"$gte": time.Now(),
			},
		}
		query[method] = identifier
		err := vc.Find(query).Sort("-createdAt").Limit(1).One(&code)
	// ... existing code ...
		_ = vc.RemoveId(code.ID)
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	if method == "phone" {
		err = c.Find(bson.M{"phone": identifier}).One(&account)
	} else if method == "email" {
		err = c.Find(bson.M{"email": identifier}).One(&account)
	}
	// ... existing code ...
	_ = c.Update(bson.M{"_id": account.ID}, bson.M{"$set": bson.M{
		"lastLoginAt": account.LastLoginAt}})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) register(method string, account *Account, verificationCode *string) (*Account, error) {
	// ... existing code ...
	if verificationCode != nil {
		c := session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
		// ... existing code ...
		query := bson.M{
			"code": *verificationCode,
			"type": "register",
			"expiresAt": bson.M{
				"$gte": time.Now(),
			},
		}
		// ... existing code ...
		err := c.Find(query).Sort("-createdAt").Limit(1).One(&code)
		// ... existing code ...
		_ = c.RemoveId(code.ID)
	}
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	if method == "phone" {
		err = c.Find(bson.M{"phone": *account.Phone}).One(&existingAccount)
	} else if method == "email" {
		err = c.Find(bson.M{"email": *account.Email}).One(&existingAccount)
	}

	if err != nil {
		if err == mgo.ErrNotFound {
			// ... existing code ...
			err = c.Insert(account)
			// ... existing code ...
		}
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (accountService *AccountService) RegisterTempAccount(account *Account) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err = c.Insert(account)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) getAccount(token string, clean bool) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"token": token}).One(&account)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) checkActivateCodeAccount(account *Account) (bool, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACTIVATE_CODE_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"userId": account.ID}).One(&accountActivateCode)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) GetAccountByPhone(phone string, clean bool) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"phone": phone}).One(&account)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) GetAccountByEmail(email string, clean bool) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"email": email}).One(&account)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) GetAccountById(userID string, clean bool) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(userID)}).One(&account)
	// ... existing code ...
	if account.ReferCode == nil {
		// ... existing code ...
		c.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$set": bson.M{"referCode": referCode}})
	}
	// ... existing code ...
}
```

```go
func (accountService *AccountService) updateAccount(token string,
	account *Account) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err = c.UpdateId(existingAccount.ID, existingAccount)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) UpdateAccountMembership(accountID bson.ObjectId, membershipExpiresAt time.Time) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	err = c.Update(bson.M{
		"_id": accountID,
	}, bson.M{
		"$set": bson.M{
			"membershipExpiresAt": membershipExpiresAt,
		},
	})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) UpdateAccountMembershipAutoRenew(accountID bson.ObjectId, autoRenew bool) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	err = c.Update(bson.M{
		"_id": accountID,
	}, bson.M{
		"$set": bson.M{
			"membershipAutoRenew": autoRenew,
		},
	})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) DeleteAccount(operatorID bson.ObjectId, accountID bson.ObjectId) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"_id": accountID}).One(&account)
	// ... existing code ...
	ca := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err = ca.Find(bson.M{"userId": accountID}).All(&connectAccounts)
	// ... existing code ...
	dc := session.DB("").C(utils.ZHIZI_DELETED_STUFF_COLLECTION)
	// ... existing code ...
	err = dc.Insert(DeletedStuff{
		// ... existing code ...
	})
	// ... existing code ...
	_, err = c.RemoveAll(bson.M{"_id": accountID})
	// ... existing code ...
	_, err = ca.RemoveAll(bson.M{"userId": accountID})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) sendCode(method string, identifier string, codeType string) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	if method == "phone" {
		count, err = c.Find(bson.M{"phone": identifier}).Count()
	} else if method == "email" {
		count, err = c.Find(bson.M{"email": identifier}).Count()
	}
	// ... existing code ...
	c = session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
	// ... existing code ...
	_, err = c.RemoveAll(query)
	// ... existing code ...
	err = c.Insert(verificationCode)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) resetPassword(method string, identifier string, verificationCode string,
	password string) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
	// ... existing code ...
	query := bson.M{
		"code": verificationCode,
		"type": "reset_password",
		"expiresAt": bson.M{
			"$gte": time.Now(),
		},
	}
	query[method] = identifier
	err := c.Find(query).Sort("-createdAt").Limit(1).One(&code)
	// ... existing code ...
	c = session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	err = c.Find(bson.M{method: identifier}).One(&account)
	// ... existing code ...
	err = c.UpdateId(account.ID, account)
	// ... existing code ...
	c = session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
	_ = c.RemoveId(code.ID)

	return cleanAccount(&account, true), nil
}
```

```go
func (accountService *AccountService) changeIdentifier(userID bson.ObjectId, method string, identifier string, verificationCode string) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		method: identifier,
		"code": verificationCode,
		"type": "change_" + method,
		"expiresAt": bson.M{
			"$gte": time.Now(),
		},
	}).Sort("-createdAt").Limit(1).One(&code)
	// ... existing code ...
	c = session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err = c.Find(bson.M{method: identifier}).One(&account)
	// ... existing code ...
	err = c.Find(bson.M{"_id": userID}).One(&account)
	// ... existing code ...
	err = c.UpdateId(account.ID, account)
	// ... existing code ...
	c = session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
	_ = c.RemoveId(code.ID)

	return cleanAccount(&account, true), nil
}
```

```go
func (accountService *AccountService) changePassword(userID string, oldPassword string,
	password string) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(userID)}).One(&account)
	// ... existing code ...
	err = c.UpdateId(account.ID, account)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) setReferrer(userID bson.ObjectId, referCode string) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err = c.Find(bson.M{"referCode": referCode}).One(&referrer)
	// ... existing code ...
	err = c.Update(bson.M{"_id": account.ID}, bson.M{"$set": bson.M{"referrerUserId": referrer.ID}})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) CreateConnectAccount(parentAccountID bson.ObjectId, account *ConnectAccount) (*ConnectAccount, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"connectUsername": account.ConnectUsername}).One(&existingAccount)

	if err != nil {
		if err == mgo.ErrNotFound {
			// ... existing code ...
			err = c.Insert(account)
			// ... existing code ...
		}
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (accountService *AccountService) ConnectAccountLogin(connectUsername string, connectPassword string) (*ConnectAccount, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"connectUsername": connectUsername}).One(&connectAccount)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) ListConnectAccounts(parentAccountID bson.ObjectId) ([]ConnectAccount, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"userId": parentAccountID}).All(&accounts)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) ListAllConnectAccounts() ([]ConnectAccount, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{}).All(&accounts)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) GetConnectAccount(parentAccountID bson.ObjectId, connectAccountID bson.ObjectId) (*ConnectAccount, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    connectAccountID,
		"userId": parentAccountID,
	}).One(&existingAccount)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) GetConnectAccountByConnectUsername(connectUsername string) (*ConnectAccount, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		"connectUsername": connectUsername,
	}).One(&existingAccount)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) UpdateConnectAccount(parentAccountID bson.ObjectId, connectAccount *ConnectAccount) (*ConnectAccount, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    connectAccount.ID,
		"userId": parentAccountID,
	}).One(&existingAccount)
	// ... existing code ...
		n, err := c.Find(bson.M{
			"connectUsername": connectAccount.ConnectUsername,
		}).Count()
	// ... existing code ...
	err = c.UpdateId(existingAccount.ID, existingAccount)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) DeleteConnectAccount(parentAccountID bson.ObjectId, connectAccountID bson.ObjectId) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_CONNECT_ACCOUNT_COLLECTION)
	err := c.Remove(bson.M{
		"userId": parentAccountID,
		"_id":    connectAccountID,
	})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) AggregateByDate(options AccountSearchOptions) (*AccountAggregateByDateResultPage, error) {
	query := accountService.constructQuery(options)
	pipeline := []bson.M{
		{
			"$match": query,
		},
		{
			"$project": bson.M{
				"yearMonthDay": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "timezone": "+08:00", "date": "$createdAt"},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$yearMonthDay",
				"totalNewUsers": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	// ... existing code ...
	err := c.Pipe(pipeline).All(&resp)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) FetchSocketIOToken(userID bson.ObjectId, request *FetchSocketIORequest) (*SocketIOToken, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_SOCKET_IO_TOKEN_COLLECTION)
	err := c.Insert(socketIOToken)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) GetIkatagoArgs(request *IKatagoArgsRequest) (*IKatagoArgsResponse, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_SOCKET_IO_TOKEN_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{"token": request.SocketIOToken, "expiredAt": bson.M{"$gte": now}}).One(&socketIOToken)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) AddRole(userID bson.ObjectId, role string) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	err = c.Update(bson.M{
		"_id": account.ID,
	}, bson.M{"$set": bson.M{
		"role":      account.Roles,
		"updatedAt": time.Now().Truncate(time.Millisecond),
	}})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) DeleteRole(userID bson.ObjectId, role string) (*Account, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	err = c.Update(bson.M{
		"_id": account.ID,
	}, bson.M{"$set": bson.M{
		"role":      account.Roles,
		"updatedAt": time.Now().Truncate(time.Millisecond),
	}})
	// ... existing code ...
}
```

```go
func (accountService *AccountService) grantMigrate(userID bson.ObjectId) (*MigrateGrantToken, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_MIGRATE_GRANT_TOKEN_COLLECTION)
	// ... existing code ...
	err = c.Insert(token)
	// ... existing code ...
}
```

```go
func (accountService *AccountService) doMigrate(userID bson.ObjectId, grantToken string) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_MIGRATE_GRANT_TOKEN_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		"grantToken": grantToken,
		"expiredAt":  bson.M{"$gte": time.Now()},
	}).One(&token)
	// ... existing code ...
	ac := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)
	err = ac.Update(bson.M{
		"_id": tempAccount.ID,
	}, bson.M{
		"$set": bson.M{
			"deleted":    true,
			"bindUserId": userID,
		},
	})
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/auth/account.go`
### Unit Tests ###
Add unit tests for:
	index = mgo.Index{
		Key:    []string{"token"},
		Unique: true,
	}
make sure that the created index is unique index

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.IKATAGO_USER_ACCOUNT_COLLECTION)

	index := mgo.Index{
		Key: []string{"phone"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"email"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"token"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"membershipExpiresAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"referCode"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"createdAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	vc := session.DB("").C(utils.IKATAGO_USER_VERIFICATION_CODES_COLLECTION)
	index = mgo.Index{
		Key: []string{"phone", "type"},
	}
	if err := vc.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"email", "type"},
	}
	if err := vc.EnsureIndex(index); err != nil {
		return err
	}
	ac := session.DB("").C(utils.IKATAGO_USER_ACTIVATE_CODE_COLLECTION)
	index = mgo.Index{
		Key:    []string{"activateCode"},
		Unique: true,
	}
	if err := ac.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"userId"},
	}
	if err := ac.EnsureIndex(index); err != nil {
		return err
	}

	st := session.DB("").C(utils.IKATAGO_SOCKET_IO_TOKEN_COLLECTION)
	index = mgo.Index{
		Key:    []string{"token"},
		Unique: true,
	}
	if err := st.EnsureIndex(index); err != nil {
		return err
	}

	return nil
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/payment/payment.go`

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_PAYMENTS_COLLECTION)

	index := mgo.Index{
		Key: []string{"userId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"userId", "-createdAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
	return nil
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/pay/service.go`

```go
func (service *Service) GetOrder(userID bson.ObjectId, orderId string) (*Order, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_PAY_ORDERS_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		"_id":    bson.ObjectIdHex(orderId),
		"userId": userID,
	}).One(&order)
	// ... existing code ...
}
```

```go
func (service *Service) CreateOrder(userID bson.ObjectId, order *Order) (*Order, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_PAY_ORDERS_COLLECTION)
	err := c.Insert(order)
	// ... existing code ...
}
```

```go
func (service *Service) HandleWechatPaidNotify(paidResult *notify.PaidResult) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_PAY_ORDERS_COLLECTION)
	cc := session.DB("").C(utils.IKATAGO_CLUSTER_USER_CREDITS_COLLECTION)

	order := Order{}
	err := c.Find(bson.M{
		"_id": bson.ObjectIdHex(*paidResult.OutTradeNo),
	}).One(&order)
	// ... existing code ...
					err = cc.Insert(&credit.CreditRecord{
						ID:         bson.NewObjectId(),
						UserID:     *account.ReferrerUserID,
						Amount:     math.Round(float64(order.Amount)*0.05) / 100,
						CreditType: credit.CREDIT_TYPE_COUPON,
						CreatedAt:  time.Now().Truncate(time.Millisecond),
						ExtraInfo: map[string]interface{}{
							"type":                "referee_for_product",
							"refereeUserId":       account.ID,
							"refereeCreditAmount": float64(order.Amount) / 100.0,
						},
					})
	// ... existing code ...
	err = c.Update(bson.M{
		"_id": order.ID,
	}, order)
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/balance/checkpoint.go`
### Unit Tests ###
add unit tests to test err := c.Find(bson.M{
		"finished": false,
	}).Limit(100).All(&jobs)
make sure the limit().All() works

Add unit tests to test the updateId with the args:
bson.M{
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
			"$inc": bson.M{
				"tryCount": 1,
			},
		}


```go
func (checkpointService *CheckpointService) fetchJobs() ([]CheckpointJob, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION)
	var jobs []CheckpointJob
	err := c.Find(bson.M{
		"finished": false,
	}).Limit(100).All(&jobs)
	// ... existing code ...
}
```

```go
func (checkpointService *CheckpointService) doJob(job *CheckpointJob) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION)
	if done {
		err := c.UpdateId(job.ID, bson.M{
			"$set": bson.M{
				"finished":  true,
				"updatedAt": time.Now(),
			},
		})
		// ... existing code ...
	} else {
		// increase the try count
		err := c.UpdateId(job.ID, bson.M{
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
			"$inc": bson.M{
				"tryCount": 1,
			},
		})
		// ... existing code ...
	}

	return nil
}
```

```go
func (checkpointService *CheckpointService) AddCheckpointJob(jobType string, userID bson.ObjectId) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION)
	// ... existing code ...
	count, err := c.Find(bson.M{
		"userId":  userID,
		"jobType": jobType,
		"$or": []bson.M{
			{"createdAt": bson.M{"$gt": now.Add(-time.Hour * 24)}},
			{"finished": false},
		},
	}).Count()
	// ... existing code ...
	err = c.Insert(&CheckpointJob{
		ID:        bson.NewObjectId(),
		UserID:    userID,
		JobType:   jobType,
		TryCount:  0,
		Finished:  false,
		CreatedAt: now,
		UpdatedAt: now,
	})
	// ... existing code ...
}
```

```go
func (checkpointService *CheckpointService) GetEarningsCheckpoint(userID bson.ObjectId) (*EarningsCheckpoint, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_EARNINGS_CHECKPOINT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		"userId": userID,
	}).One(result)
	// ... existing code ...
}
```

```go
func (checkpointService *CheckpointService) GetBalanceCheckpoint(userID bson.ObjectId) (*BalanceCheckpoint, error) {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_BALANCE_CHECKPOINT_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{
		"userId": userID,
	}).One(result)
	// ... existing code ...
}
```

```go
func (checkpointService *CheckpointService) createBalanceCheckpoint(userID bson.ObjectId) (*BalanceCheckpoint, error) {
	// ... existing code ...
	uc := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	// ... existing code ...
	err = uc.Find(bson.M{
		"connectUserId": userID,
		"finished":      true,
		"serialId": bson.M{
			"$gt": lastCheckpoint.SerialID,
		},
	}).Select(bson.M{"serialId": 1}).Sort("-serialId").Limit(1).One(&lastFinishedUsage)
	// ... existing code ...
	err = uc.Find(bson.M{
		"connectUserId": userID,
		"finished":      false,
		"serialId": bson.M{
			"$gt": lastCheckpoint.SerialID,
			"$lt": maxSerialID,
		},
	}).Select(bson.M{"serialId": 1}).Sort("serialId").Limit(1).One(&minUnfinishedUsage)
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_BALANCE_CHECKPOINT_COLLECTION)
	_, err = c.Upsert(bson.M{
		"userId": userID,
	}, lastCheckpoint)
	// ... existing code ...
}
```

```go
func (checkpointService *CheckpointService) createEarningsCheckpoint(userID bson.ObjectId) (*EarningsCheckpoint, error) {
	// ... existing code ...
	uc := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	// ... existing code ...
	err = uc.Find(bson.M{
		"nodeOwnerUserId": userID,
		"finished":        true,
		"serialId": bson.M{
			"$gt": lastCheckpoint.SerialID,
		},
	}).Select(bson.M{"serialId": 1}).Sort("-serialId").Limit(1).One(&lastFinishedUsage)
	// ... existing code ...
	err = uc.Find(bson.M{
		"nodeOwnerUserId": userID,
		"finished":        false,
		"serialId": bson.M{
			"$gt": lastCheckpoint.SerialID,
			"$lt": maxSerialID,
		},
	}).Select(bson.M{"serialId": 1}).Sort("serialId").Limit(1).One(&minUnfinishedUsage)
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_EARNINGS_CHECKPOINT_COLLECTION)
	_, err = c.Upsert(bson.M{
		"userId": userID,
	}, lastCheckpoint)
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/balance/balance.go`

```go
func EnsureIndexes() error {
	session := utils.NewDBSession()
	defer session.Close()
	ecc := session.DB("").C(utils.IKATAGO_CLUSTER_EARNINGS_CHECKPOINT_COLLECTION)
	index := mgo.Index{
		Key:    []string{"userId"},
		Unique: true,
	}
	if err := ecc.EnsureIndex(index); err != nil {
		return err
	}
	bcc := session.DB("").C(utils.IKATAGO_CLUSTER_BALANCE_CHECKPOINT_COLLECTION)
	index = mgo.Index{
		Key:    []string{"userId"},
		Unique: true,
	}
	if err := bcc.EnsureIndex(index); err != nil {
		return err
	}
	cjc := session.DB("").C(utils.IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION)
	index = mgo.Index{
		Key: []string{"userId", "jobType"},
	}
	if err := cjc.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"userId", "jobType", "finished"},
	}
	if err := cjc.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"userId", "jobType", "createdAt"},
	}
	if err := cjc.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"finished"},
	}
	if err := cjc.EnsureIndex(index); err != nil {
		return err
	}
	return nil
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/gamerecord/gamerecord.go`

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_GAME_RECORDS_COLLECTION)
	ac := session.DB("").C(utils.IKATAGO_CLUSTER_USER_ANALYSIS_RECORDS_COLLECTION)
	index := mgo.Index{
		Key: []string{"userId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"userId", "-updatedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"userId", "type", "-updatedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"userId", "favorite", "-updatedAt"},
	}

	index = mgo.Index{
		Key: []string{"gameRecordId"},
	}
	if err := ac.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key:    []string{"gameRecordId", "moveId"},
		Unique: true,
	}
	if err := ac.EnsureIndex(index); err != nil {
		return err
	}
	return nil
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/usage/service.go`
### Unit Tests ###
1. Add unit tests to test the "$in" with array ids works.
query := bson.M{
		"finished": false,
		"commandIds": bson.M{
			"$in": commandIds,
		},
	}
2. add unit tests for:
index = mgo.Index{
		Key:  []string{"connectUserId", "finished", "-endedAt", "-startedAt"},
		Name: "usage_list_user_sort_index",
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
to thest the created index is with the same name.

```go
func (service *Service) loadSerialID() error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	// ... existing code ...
	err := c.Find(bson.M{}).Sort("-serialId").Limit(1).One(&lastUsage)
	// ... existing code ...
}
```

```go
func (service *Service) reloadRunningUsages() error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	usages := make([]*Usage, 0)
	err := c.Find(bson.M{"finished": false}).All(&usages)
	// ... existing code ...
}
```

```go
func (service *Service) markFinishedWithoutLock(connectUserID *bson.ObjectId, usageIds []bson.ObjectId) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	query := bson.M{
		"_id": bson.M{
			"$in": usageIds,
		},
	}
	if connectUserID != nil {
		query["connectUserId"] = connectUserID
	}
	usages := make([]Usage, 0)
	err := c.Find(query).All(&usages)
	// ... existing code ...
	for i := range usages {
		// ... existing code ...
		err := c.Update(bson.M{
			"_id": usage.ID,
		}, usage)
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (service *Service) MarkFinishedByCommandIds(userID *bson.ObjectId, commandIds []string) error {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	query := bson.M{
		"finished": false,
		"commandIds": bson.M{
			"$in": commandIds,
		},
	}
	usages := make([]Usage, 0)
	err := c.Find(query).All(&usages)
	// ... existing code ...
}
```

```go
func (service *Service) handleKatagoUsages(nodeID bson.ObjectId, nodeData *workernode.NodeData) error {
	// ... existing code ...
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)

	// ... existing code ...
		query := bson.M{
			"connectUsername": connectUsername,
			"workerId":        workerID,
			"finished":        true,
			"commandIds": bson.M{
				"$in": currentState.CommandIDs,
			},
		}
		existingFinishedUsages := make([]Usage, 0)
		err := c.Find(query).All(&existingFinishedUsages)
	// ... existing code ...
	for _, updatedUsage := range updatedUsages {
		// update the usage in db
		err := c.Update(bson.M{"_id": updatedUsage.ID}, updatedUsage)
		// ... existing code ...
	}
	for _, insertedUsage := range insertedUsages {
		// insert the usage in db
		// ... existing code ...
		err := c.Insert(insertedUsage)
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (service *Service) handleSystemKatagoUsages(nodeID bson.ObjectId, nodeData *workernode.NodeData) error {
	// ... existing code ...
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)

	// ... existing code ...
		query := bson.M{
			"connectUsername": connectUsername,
			"workerId":        workerID,
			"finished":        true,
			"commandIds": bson.M{
				"$in": currentState.CommandIDs,
			},
		}
		existingFinishedUsages := make([]Usage, 0)
		err := c.Find(query).All(&existingFinishedUsages)
	// ... existing code ...
	for _, updatedUsage := range updatedUsages {
		// update the usage in db
		err := c.Update(bson.M{"_id": updatedUsage.ID}, updatedUsage)
		// ... existing code ...
	}
	for _, insertedUsage := range insertedUsages {
		// insert the usage in db
		err := c.Insert(insertedUsage)
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (service *Service) handlePlayKatagoUsages(nodeID bson.ObjectId, nodeData *workernode.NodeData) error {
	// ... existing code ...
	session := utils.NewDBSession()
	defer session.Close()
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)

	// ... existing code ...
		query := bson.M{
			"connectUsername": connectUsername,
			"workerId":        workerID,
			"finished":        true,
			"commandIds": bson.M{
				"$in": currentState.CommandIDs,
			},
		}
		existingFinishedUsages := make([]Usage, 0)
		err := c.Find(query).All(&existingFinishedUsages)
	// ... existing code ...
	for _, updatedUsage := range updatedUsages {
		// update the usage in db
		err := c.Update(bson.M{"_id": updatedUsage.ID}, updatedUsage)
		// ... existing code ...
	}
	for _, insertedUsage := range insertedUsages {
		// insert the usage in db
		err := c.Insert(insertedUsage)
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (service *Service) calculateGamePreviousTotalCost(connectUserID bson.ObjectId, gameID string) (float64, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"connectUserId": connectUserID,
				"gameId":        gameID,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"previousTotalCost": bson.M{
					"$sum": "$totalCost",
				},
			},
		},
	}

	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	resp := []bson.M{}
	err := c.Pipe(pipeline).All(&resp)
	// ... existing code ...
}
```

```go
func (service *Service) handleDeadNodeUsages() {
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)

	for _, updatedUsage := range updatedUsages {
		// update the usage in db
		err := c.Update(bson.M{"_id": updatedUsage.ID}, updatedUsage)
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (service *Service) Search(options UsageSearchOptions) (*UsageSearchResultPage, error) {

	query := service.constructQuery(options)
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	// ... existing code ...
	count, err := c.Find(query).Count()
	// ... existing code ...
	if pageSize != nil && page != nil {
		err = c.Find(query).Sort("finished", "-endedAt", "-startedAt").Skip(*pageSize * *page).Limit(*pageSize).All(&usages)
		// ... existing code ...
	} else {
		err = c.Find(query).Sort("finished", "-endedAt", "-startedAt").All(&usages)
		// ... existing code ...
	}
	// ... existing code ...
}
```

```go
func (service *Service) Aggregate(options UsageSearchOptions) (*UsageAggregateResult, error) {
	query := service.constructQuery(options)
	pipeline := []bson.M{
		{
			"$match": query,
		},
		{
			"$group": bson.M{
				"_id": nil,
				"totalConsumption": bson.M{
					"$sum": "$totalCost",
				},
				"totalCouponConsumption": bson.M{
					"$sum": "$totalCouponCost",
				},
				"totalDuration": bson.M{
					"$sum": "$duration",
				},
				"currentNumOfMyConnections": bson.M{
					"$sum": "$numOfMyConnections",
				},
				"currentNumOfNodes": bson.M{
					"$sum": 1,
				},
				"maxSerialId": bson.M{
					"$max": "$serialId",
				},
			},
		},
	}

	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	resp := []bson.M{}
	err := c.Pipe(pipeline).All(&resp)
	// ... existing code ...
}
```

```go
func (service *Service) AggregateByDate(options UsageSearchOptions) (*UsageAggregateByDateResultPage, error) {
	query := service.constructQuery(options)
	pipeline := []bson.M{
		{
			"$match": query,
		},
		{
			"$project": bson.M{
				"yearMonthDay": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "timezone": "+08:00", "date": "$startedAt"},
				},
				"totalCost":        1,
				"duration":         1,
				"totalCouponCost":  1,
				"virtualTotalCost": 1,
			},
		},
		{
			"$group": bson.M{
				"_id": "$yearMonthDay",
				"totalConsumption": bson.M{
					"$sum": "$totalCost",
				},
				"totalVirtualConsumption": bson.M{
					"$sum": "$virtualTotalCost",
				},
				"totalCouponConsumption": bson.M{
					"$sum": "$totalCouponCost",
				},
				"totalDuration": bson.M{
					"$sum": "$duration",
				},
			},
		},
	}
	// ... existing code ...
	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	resp := []bson.M{}
	err := c.Pipe(pipeline).All(&resp)
	// ... existing code ...
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/usage/usage.go`

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	c := session.DB("").C(utils.IKATAGO_CLUSTER_USER_USAGES_COLLECTION)

	index := mgo.Index{
		Key: []string{"nodeId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"nodeOwnerUserId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
index = mgo.Index{
		Key: []string{"nodename"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
index = mgo.Index{
		Key: []string{"connectUserId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"connectUsername"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"connectUsername", "finished"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"connectUserId", "finished"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"finished"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"serialId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"connectUserId", "serialId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"connectUserId", "startedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"nodeOwnerUserId", "serialId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
index = mgo.Index{
		Key: []string{"-startedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
index = mgo.Index{
		Key: []string{"-lastUpdatedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key:  []string{"finished", "-endedAt", "-startedAt"},
		Name: "usage_list_sort_index",
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
index = mgo.Index{
		Key:  []string{"connectUserId", "finished", "-endedAt", "-startedAt"},
		Name: "usage_list_user_sort_index",
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key:  []string{"connectUsername", "finished", "commandIds"},
		Name: "usage_by_command_ids",
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}
index = mgo.Index{
		Key:  []string{"commandIds"},
		Name: "usage_simple_command_ids",
	}

	if err := c.EnsureIndex(index); err != nil {
		return err
	}

index = mgo.Index{
		Key: []string{"connectUserId", "gameId"},
	}
	if err := c.EnsureIndex(index); err != nil {
		return err
	}

	return nil
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/utils/db.go`

```go
// InitDB initializes the global database session
func InitDB() error {
	log.Infof("connecting to db: %s", config.GetConfig().GetString("mongodb.url"))
	session, err := mgo.DialModernMGO(config.GetConfig().GetString("mongodb.url"))
	if err != nil {
		return err
	}

	session.SetMode(mgo.Monotonic, true)

	gSession = session

	return nil
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/team/team.go`

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	tc := session.DB("").C(utils.IKATAGO_CLUSTER_TEAM_COLLECTION)

	index := mgo.Index{
		Key: []string{"adminUserId"},
	}

	if err := tc.EnsureIndex(index); err != nil {
		return err
	}

	crc := session.DB("").C(utils.IKATAGO_CLUSTER_TEAM_CHARGE_RECORD_COLLECTION)
	index = mgo.Index{
		Key: []string{"teamId"},
	}

	if err := crc.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"teamId", "-createdAt"},
	}

	if err := crc.EnsureIndex(index); err != nil {
		return err
	}

	trc := session.DB("").C(utils.IKATAGO_CLUSTER_TEAM_TRANSFER_RECORD_COLLECTION)
	index = mgo.Index{
		Key: []string{"teamId"},
	}

	if err := trc.EnsureIndex(index); err != nil {
		return err
	}

	index = mgo.Index{
		Key: []string{"teamId", "-createdAt"},
	}

	if err := trc.EnsureIndex(index); err != nil {
		return err
	}
	return nil
}
```

---

### `/Users/jinggangwang/gochess/ikatago-service/giftcard/giftcard.go`

```go
// EnsureIndices make the indices
func EnsureIndices() error {
	session := utils.NewDBSession()
	defer session.Close()

	gcc := session.DB("").C(utils.IKATAGO_CLUSTER_GIFTCARD_COLLECTION)

	index := mgo.Index{
		Key:    []string{"giftCardCode"},
		Unique: true,
	}

	if err := gcc.EnsureIndex(index); err != nil {
		return err
	}
	index = mgo.Index{
		Key: []string{"createUserId", "-createdAt"},
	}

	if err := gcc.EnsureIndex(index); err != nil {
		return err
	}

	return nil
}
```
