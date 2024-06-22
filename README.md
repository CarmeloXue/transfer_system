# transfer_system
A transfer system to allow account transfer between each other

# How to start


# System Design

Below is the system diagram of the transfer system. I have:

1. Nginx: This layer works as a reverse proxy to find corresponding upstream services. 
2. Business application: There just one logical container(we can separate them, and make it easy to scale too) here, on this http server, I provided two services
   1. AccountService: To handle account creation and query as well as account balance management. It will provide TCC functions to handle balance change in it's own transaction.
   2. TransactionService: To hanlder transaction between accounts. It's more like a coordinator to trigger TCC functions from AccountService.
3. Database: This layer used to store accounts and transactions. 

![Transfer System Architecture](/diagram.png)


## Databases

I have two separate databases here, one for AccountService and one for TransactionService. The reason I put them in separate db is because in real world application, we may have db shardings, and we may not be able to do all the actions within one transaction within one db.

1. `account_db`. It's used for AccountService, containing two tables:
   1. `account_tab`. This table contains registered user's account information including balance.
   2. `fund_movement-tab`. This table will record fund movement for a account and link the fund movement to it's parent transaction in transaction_db
2. `transaction_db`. It's used for TransactionService to move the state of a fund transfer between two accounts.
   1. `transaction_tab`. It's only table in this db. Recording all the transactions. Each transaction here will have(and only) two fund movement records in account_db.
   

## API

### Account service

##### AccountQuery

```
 - Methods: Get
 - URL: /api/v1/account/{account_id}
 - Response: 
 {
    "success": bool,
    "error_code": uint32,
    "error_message": string,
    "data": {
        "account": uint64,
        "balance": string,
    }
 }
```

##### AccountCreation

```
 - Methods: Post
 - URL: /api/v1/account
 - Request body: 
 {
    "account_id": uint64,
    "initial_balance": string
 }

 - Response: 
 {
    "success": bool,
    "error_code": uint32,
    "error_message": string,
    "data": {
        "account_id": uint64,
        "balance": string,
    }
 }
```


### Transaction Service

##### Submit transfer


```
 - Methods: Post
 - URL: /api/v1/payment
 - Request body: 
 {
    "from": uint64,
    "to": uint64,
    "amount": float64,
 }
 
 - Response: 
 {
    "success": bool,
    "error_code": uint32,
    "error_message": string,
    "data": {
        "order_id": uint64,
        "order_status": string,
        "order_amount": float64,
        "from": uint64,
        "to": uint64
    }
 }
```