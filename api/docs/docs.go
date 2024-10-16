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
                            "$ref": "#/definitions/responses.VerificationStatusResponse"
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
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/responses.TokenResponseWithStatus"
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
        "responses.TokenResponseWithStatus": {
            "type": "object",
            "properties": {
                "is_new_token": {
                    "type": "boolean"
                },
                "message": {
                    "type": "string"
                },
                "token": {
                    "$ref": "#/definitions/responses.TokenResponse"
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
                "scanRef": {
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
                "autoDocument": {
                    "type": "string"
                },
                "autoFace": {
                    "type": "string"
                },
                "clientId": {
                    "type": "string"
                },
                "fraudTags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "manualDocument": {
                    "type": "string"
                },
                "manualFace": {
                    "type": "string"
                },
                "mismatchTags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "scanRef": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
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
