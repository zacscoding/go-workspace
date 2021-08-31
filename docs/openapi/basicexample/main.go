package main

import (
	"context"
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func main() {
	fp, err := filepath.Abs("./docs/openapi/basicexample/spec.yaml")
	if err != nil {
		log.Fatalf("failed to get absolute path. err: %v", err)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(fp)
	if err != nil {
		log.Fatalf("failed to load api spec. err: %v", err)
	}

	if err := doc.Validate(context.Background()); err != nil {
		log.Fatalf("failed to validate docs. err: %v", err)
	}

	for k, path := range doc.Paths {
		log.Println("//---------------------------------------------------------------")
		log.Printf("## Check %s", k)
		log.Println("//---------------------------------------------------------------")

		// Check parameters
		//---------------------------------------------------------------
		log.Println("//------------------------------")
		log.Printf("### Check parameters")
		for _, param := range path.Parameters {
			log.Printf("> Ref: %s", param.Ref)
			log.Printf("> %s", mustToJson(param.Value))
		}
		log.Println()

		ops := map[string]*openapi3.Operation{
			http.MethodGet:    path.Get,
			http.MethodPost:   path.Post,
			http.MethodPut:    path.Put,
			http.MethodDelete: path.Delete,
		}

		for m, op := range ops {
			log.Println("//------------------------------")
			log.Printf("## Check Operations. method: %s", m)
			if op == nil {
				log.Println("> Empty operation")
				log.Println("//------------------------------")
				continue
			}

			log.Printf("## Check parameters")
			for _, param := range op.Parameters {
				log.Printf("> Ref: %s", param.Ref)
				log.Printf("> %s", mustToJson(param.Value))
			}
			log.Println()

			log.Printf("## Check security")
			if op.Security == nil {
				log.Println("> Empty security")
			} else {
				for _, requirement := range *op.Security {
					for key, values := range requirement {
						log.Printf("> Key: %s, Values: %s", key, strings.Join(values, ","))
					}
				}
			}
			log.Println()

			log.Printf("## Check response")
			for key, res := range op.Responses {
				log.Printf("> Key: %s, Res: %s", key, mustToJson(res))
			}

			log.Println("//------------------------------")
		}
	}

	log.Println("//---------------------------------------------------------------")
	log.Println("## Check components")
	log.Println("//---------------------------------------------------------------")
	for key, ref := range doc.Components.Schemas {
		log.Printf("# %s", key)
		log.Printf("> Ref: %s", ref.Ref)
	}
}

func mustToJson(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
