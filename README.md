# transfer_system
A transfer system to allow account transfer between each other

# How to start


# System Design

Below is the system diagram of the transfer system. I have:

1. Nginx: This layer works as a reverse proxy to find corresponding upstream services. 
2. Business application: This layer contains two services
   1. AccountService: To handle account creation and query
   2. TransactionService: To Process User's Transaction
3. Database: This layer used to store accounts and transactions

![Transfer System Architecture](/diagram.png)


## Databases

I put one database `transfer_db` and keep all tables inside this database. Inside I designed 3 tables. They are: 

1. `account_tab`. This table will holds registered account information including id and balance information.
2. `order_tab`. An order represents a transfer from one account to another. It will start from `pending` to `success`. And each successful order(transfer) will have two transactions.
3. `transaction_tab`. An transaction describes a monetary movement to a user's account. It just one direction, either `payment` or `payment_recieved`. Each transaction will have a corresponding parent order.
   

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
        "balance": float64,
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
    "balance": float64
 }

 - Response: 
 {
    "success": bool,
    "error_code": uint32,
    "error_message": string,
    "data": {
        "account": uint64,
        "balance": float64,
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