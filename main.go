package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ShankaranarayananBR/lambda-base/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

var logger *zap.Logger

var db *sql.DB

func init() {
	l, _ := zap.NewProduction()
	logger = l

	dbConnection, err := database.GetConnection()
	if err != nil {
		logger.Error("error connecting to the database", zap.Error(err))
		panic(err)
	}
	dbConnection.Ping()
	if err != nil {
		logger.Error("error pinging the database", zap.Error(err))
		panic(err)
	}
	db = dbConnection
}

type DefaultResponse struct {
	Status  string `json:"status`
	Message string `json:"message"`
}

type GetEmployeesResponse struct {
	Employees []*database.Employee `json:"employees"`
}

func MyHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var resp *events.APIGatewayProxyResponse
	logger.Info("received Event", zap.Any("method", event.HTTPMethod), zap.Any("path", event.Path), zap.Any("body", event.Body))
	if event.Path == "/migrate" {
		body, _ := json.Marshal(&DefaultResponse{
			Status:  string(http.StatusOK),
			Message: "Hello World!",
		})
		err := database.CreateEmployeesTable(ctx, db)

		if err != nil {
			body, _ := json.Marshal(&DefaultResponse{
				Status:  string(http.StatusOK),
				Message: "could not create employees table",
			})

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(body),
			}, nil
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(body),
		}, nil
	} else if event.Path == "/employees" && event.HTTPMethod == http.MethodPost {
		//create a new employee
		employee := &database.Employee{}
		err := json.Unmarshal([]byte(event.Body), &employee)
		if err != nil {
			body, _ := json.Marshal(&DefaultResponse{
				Status:  string(http.StatusBadRequest),
				Message: err.Error(),
			})

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       string(body),
			}, nil
		}

		err = database.CreateEmployee(ctx, db, employee.Email, employee.FirstName, employee.LastName)
		if err != nil {
			body, _ := json.Marshal(&DefaultResponse{
				Status:  string(http.StatusInternalServerError),
				Message: err.Error(),
			})
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       string(body),
			}, nil
		}
		body, _ := json.Marshal(&DefaultResponse{
			Status:  string(http.StatusOK),
			Message: err.Error(),
		})
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(body),
		}, nil

	} else if event.Path == "/employees" && event.HTTPMethod == http.MethodGet {
		// get all employees
		employees, err := database.GetEmployees(ctx, db)
		if err != nil {
			body, _ := json.Marshal(&DefaultResponse{
				Status:  string(http.StatusInternalServerError),
				Message: "error fetching employees",
			})

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       string(body),
			}, nil
		}

		body, _ := json.Marshal(&GetEmployeesResponse{
			Employees: employees,
		})

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       string(body),
		}, nil
	} else {
		body, _ := json.Marshal(&DefaultResponse{
			Status:  string(http.StatusOK),
			Message: "default endpoint",
		})
		resp = &events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(body),
		}

	}

	return *resp, nil
}

func main() {
	lambda.Start(MyHandler)
}
