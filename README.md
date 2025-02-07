# Account Service API

This project implements a RESTful API for a simple banking account service, built with Go, Fiber, PostgreSQL, and Docker. It provides endpoints for registering customers, depositing funds, withdrawing funds, and checking account balances.

## Features

*   **Customer Registration (`/daftar`):**
    *   Allows new customers to register with their name, national ID number (NIK), and mobile phone number.
    *   Generates a unique account number upon successful registration.
    *   Performs validation to prevent duplicate NIK and phone numbers.
    *   Returns a JSON response with the new account number.
*   **Deposit (`/tabung`):**
    *   Allows registered customers to deposit funds into their accounts.
    *   Requires the account number and deposit amount.
    *   Updates the account balance and records the transaction.
    *   Returns the updated account balance.
*   **Withdrawal (`/tarik`):**
    *   Allows registered customers to withdraw funds from their accounts.
    *   Requires the account number and withdrawal amount.
    *   Checks for sufficient balance before processing the withdrawal.
    *   Updates the account balance and records the transaction.
    *   Returns the updated account balance.
*   **Balance Inquiry (`/saldo/{no_rekening}`):**
    *   Allows customers to check their account balance.
    *   Requires the account number as a path parameter.
    *   Returns the current account balance.
* **Transaction History (Chained):**
    * Each deposit and withdrawal creates a `cash_activity` record.
    * Uses `reference_id` to make a chained transaction history.
* **Comprehensive Unit and Integration Tests:**
	* Unit tests are implemented by using mock db.
	* Integration tests are implemented by running the app and test with real database connection.
* **Structured Logging:** uses logrus.
* **Swagger Documentation:** API documentation auto generated.

## API Endpoints

| Method | Endpoint            | Description                                      | Request Body                                    | Success Response (200/201)                      | Error Responses                                                                           |
| ------ | ------------------- | ------------------------------------------------ | ----------------------------------------------- | ------------------------------------------------ | ---------------------------------------------------------------------------------------- |
| POST   | `/daftar`           | Register a new customer.                        | `{ "nama": "string", "nik": "string", "no_hp": "string" }` | `{ "code": 201, "status": "success", "message":"Account registration successful", "data": { "account_number": "string" } }`               | 400 (Bad Request - validation errors), 409 (Conflict - duplicate NIK/phone)           |
| POST   | `/tabung`          | Deposit funds into an account.                  | `{ "no_rekening": "string", "nominal": number }` | `{ "code": 200, "status": "success", "message":"Deposit successful", "data": number (balance) }`        | 400 (Bad Request - validation), 404 (Not Found - account doesn't exist)                |
| POST   | `/tarik`           | Withdraw funds from an account.                 | `{ "no_rekening": "string", "nominal": number }` |  `{ "code": 200, "status": "success", "message":"Withdrawal successful", "data": number(balance) }`       | 400 (Bad Request - validation/insufficient balance), 404 (Not Found - account)      |
| GET    | `/saldo/{no_rekening}` | Get the balance of an account.                | *None*                                          | `{ "code": 200, "status": "success", "message": "Get balance successful", "data": number (balance) }` | 400 (Bad Request - invalid account number format), 404 (Not Found - account) |

## Technology Stack
- Go: Programming language.
- Fiber: Web framework.
- PostgreSQL: Relational database.
- GORM: ORM library for interacting with the database.
- Docker: Containerization.
- Docker Compose: Multi-container orchestration.
- golang-migrate/migrate: Database migrations.
- swaggo/swag: Swagger documentation generation.
- testify: for assertion in unit testing.
- logrus: for structured logging.
- go-playground/validator/v10: for data validation.




## Project Structure
```bash
account-service/
├── src/                # Source code
│   ├── controller/     # API handlers
│   ├── database/       # Database connection setup
│   ├── model/          # Data models (structs)
│   ├── service/        # Business logic
│   ├── validation/     # Request validation structs
│   ├── main.go         # Main application entry point
│   ├── go.mod
│   ├── go.sum
│   └── docs/           # Swagger documentation
├── test/
│   ├── fixture/        # Example of data
│   ├── helper/         # Helper function for testing
│   ├── integration/    # Integration tests
│   └── unit/           # Unit tests
├── Dockerfile          # Dockerfile for building the service
├── docker-compose.yml  # Docker Compose file for deployment
└── docker-compose.test.yml # Docker Compose override for testing
└── entrypoint.sh       # Entrypoint Script
```

Setup and Deployment
Environment Variables:

Create a .env file in the project root. See .env.example.
Or, set the required environment variables directly in your shell.
Build and Run with Docker Compose (Recommended):
```bash
docker compose up --build -d
```
This command builds the Docker image (including running unit tests), starts the PostgreSQL and account-service containers, and applies database migrations.

Build and Run without Docker Compose:
```bash
go run src/main.go -port=3000 -host=localhost
```


Run migrations: 
```bash
make migrate-docker-up
```


Run unit tests:

```bash
make tests
```


Access the API:

The API will be available at http://localhost:3000 (or the port you configured).
Swagger UI: http://localhost:3000/swagger/index.html


## ERD


## License

[MIT](https://choosealicense.com/licenses/mit/)