// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "Codescalers Egypt",
            "url": "https://codescalers-egypt.com",
            "email": "info@codescalers.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/configs": {
            "get": {
                "description": "Returns the service configs",
                "tags": [
                    "Misc"
                ],
                "summary": "Get Service Configs",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {}
                    }
                }
            }
        },
        "/api/v1/data": {
            "get": {
                "description": "Returns the verification data for a client",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Verification"
                ],
                "summary": "Get Verification Data",
                "parameters": [
                    {
                        "maxLength": 48,
                        "minLength": 48,
                        "type": "string",
                        "description": "TFChain SS58Address",
                        "name": "X-Client-ID",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "hex-encoded message ` + "`" + `{api-domain}:{timestamp}` + "`" + `",
                        "name": "X-Challenge",
                        "in": "header",
                        "required": true
                    },
                    {
                        "maxLength": 128,
                        "minLength": 128,
                        "type": "string",
                        "description": "hex-encoded sr25519|ed25519 signature",
                        "name": "X-Signature",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.VerificationDataResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/health": {
            "get": {
                "description": "Returns the health status of the service",
                "tags": [
                    "Health"
                ],
                "summary": "Health Check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.HealthResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/status": {
            "get": {
                "description": "Returns the verification status for a client",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Verification"
                ],
                "summary": "Get Verification Status",
                "parameters": [
                    {
                        "maxLength": 48,
                        "minLength": 48,
                        "type": "string",
                        "description": "TFChain SS58Address",
                        "name": "client_id",
                        "in": "query"
                    },
                    {
                        "minLength": 1,
                        "type": "string",
                        "description": "Twin ID",
                        "name": "twin_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.VerificationStatusResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/token": {
            "post": {
                "description": "Returns a token for a client",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Token"
                ],
                "summary": "Get or Generate iDenfy Verification Token",
                "parameters": [
                    {
                        "maxLength": 48,
                        "minLength": 48,
                        "type": "string",
                        "description": "TFChain SS58Address",
                        "name": "X-Client-ID",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "hex-encoded message ` + "`" + `{api-domain}:{timestamp}` + "`" + `",
                        "name": "X-Challenge",
                        "in": "header",
                        "required": true
                    },
                    {
                        "maxLength": 128,
                        "minLength": 128,
                        "type": "string",
                        "description": "hex-encoded sr25519|ed25519 signature",
                        "name": "X-Signature",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Existing token retrieved",
                        "schema": {
                            "$ref": "#/definitions/responses.TokenResponse"
                        }
                    },
                    "201": {
                        "description": "New token created",
                        "schema": {
                            "$ref": "#/definitions/responses.TokenResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "402": {
                        "description": "Payment Required",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/responses.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/version": {
            "get": {
                "description": "Returns the service version",
                "tags": [
                    "Misc"
                ],
                "summary": "Get Service Version",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/webhooks/idenfy/id-expiration": {
            "post": {
                "description": "Processes the doc expiration notification for a client",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Webhooks"
                ],
                "summary": "Process Doc Expiration Notification",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/webhooks/idenfy/verification-update": {
            "post": {
                "description": "Processes the verification update for a client",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Webhooks"
                ],
                "summary": "Process Verification Update",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        }
    },
    "definitions": {
        "responses.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "responses.HealthResponse": {
            "type": "object",
            "properties": {
                "errors": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "status": {
                    "$ref": "#/definitions/responses.HealthStatus"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "responses.HealthStatus": {
            "type": "string",
            "enum": [
                "Healthy",
                "Degraded"
            ],
            "x-enum-varnames": [
                "HealthStatusHealthy",
                "HealthStatusDegraded"
            ]
        },
        "responses.Outcome": {
            "type": "string",
            "enum": [
                "VERIFIED",
                "REJECTED"
            ],
            "x-enum-varnames": [
                "OutcomeVerified",
                "OutcomeRejected"
            ]
        },
        "responses.TokenResponse": {
            "type": "object",
            "properties": {
                "authToken": {
                    "type": "string"
                },
                "clientId": {
                    "type": "string"
                },
                "digitString": {
                    "type": "string"
                },
                "expiryTime": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                },
                "scanRef": {
                    "type": "string"
                },
                "sessionLength": {
                    "type": "integer"
                },
                "tokenType": {
                    "type": "string"
                }
            }
        },
        "responses.VerificationDataResponse": {
            "type": "object",
            "properties": {
                "additionalData": {},
                "address": {
                    "type": "string"
                },
                "addressVerification": {},
                "ageEstimate": {
                    "type": "string"
                },
                "authority": {
                    "type": "string"
                },
                "birthPlace": {
                    "type": "string"
                },
                "clientId": {
                    "type": "string"
                },
                "clientIpProxyRiskLevel": {
                    "type": "string"
                },
                "docBirthName": {
                    "type": "string"
                },
                "docDateOfIssue": {
                    "type": "string"
                },
                "docDob": {
                    "type": "string"
                },
                "docExpiry": {
                    "type": "string"
                },
                "docFirstName": {
                    "type": "string"
                },
                "docIssuingCountry": {
                    "type": "string"
                },
                "docLastName": {
                    "type": "string"
                },
                "docNationality": {
                    "type": "string"
                },
                "docNumber": {
                    "type": "string"
                },
                "docPersonalCode": {
                    "type": "string"
                },
                "docSex": {
                    "type": "string"
                },
                "docTemporaryAddress": {
                    "type": "string"
                },
                "docType": {
                    "type": "string"
                },
                "driverLicenseCategory": {
                    "type": "string"
                },
                "duplicateDocFaces": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "duplicateFaces": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "fullName": {
                    "type": "string"
                },
                "idenfyRef": {
                    "type": "string"
                },
                "manuallyDataChanged": {
                    "type": "boolean"
                },
                "mothersMaidenName": {
                    "type": "string"
                },
                "orgAddress": {
                    "type": "string"
                },
                "orgAuthority": {
                    "type": "string"
                },
                "orgBirthName": {
                    "type": "string"
                },
                "orgBirthPlace": {
                    "type": "string"
                },
                "orgFirstName": {
                    "type": "string"
                },
                "orgLastName": {
                    "type": "string"
                },
                "orgMothersMaidenName": {
                    "type": "string"
                },
                "orgNationality": {
                    "type": "string"
                },
                "orgTemporaryAddress": {
                    "type": "string"
                },
                "selectedCountry": {
                    "type": "string"
                }
            }
        },
        "responses.VerificationStatusResponse": {
            "type": "object",
            "properties": {
                "clientId": {
                    "type": "string"
                },
                "final": {
                    "type": "boolean"
                },
                "idenfyRef": {
                    "type": "string"
                },
                "status": {
                    "$ref": "#/definitions/responses.Outcome"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1.0",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "TFGrid KYC API",
	Description:      "This is a KYC service for TFGrid.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
