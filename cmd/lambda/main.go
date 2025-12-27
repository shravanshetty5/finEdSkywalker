package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sshetty/finEdSkywalker/internal/handlers"
)

func main() {
	log.Println("Starting Lambda function...")
	lambda.Start(handlers.Handler)
}

