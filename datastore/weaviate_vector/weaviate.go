package weaviateVS

import (
	"context"
	"errors"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/wejick/gochain/datastore"
	"github.com/wejick/gochain/model"
)

var _ datastore.VectorStore = &WeaviateVectorStore{}

type WeaviateVectorStore struct {
	client         *weaviate.Client
	embeddingModel model.EmbeddingModel

	existClass map[string]bool
}

func NewWeaviateVectorStore(host string, scheme string, apiKey string, embeddingModel model.EmbeddingModel, headers map[string]string) (WVS *WeaviateVectorStore, err error) {
	WVS = &WeaviateVectorStore{
		existClass:     map[string]bool{},
		embeddingModel: embeddingModel,
	}
	cfg := weaviate.Config{
		Host:       host,
		Scheme:     scheme,
		Headers:    headers,
		AuthConfig: auth.ApiKey{Value: apiKey},
	}
	WVS.client, err = weaviate.NewClient(cfg)

	return
}

func (W *WeaviateVectorStore) SearchVector(ctx context.Context, className string, vector []float32) (output []interface{}, err error) {
	query := W.client.GraphQL().NearVectorArgBuilder().WithVector(vector)
	fields := []graphql.Field{
		{Name: "text"},
	}
	resp, err := W.client.GraphQL().Get().WithClassName(className).WithNearVector(query).WithFields(fields...).Do(ctx)
	if err != nil {
		return
	}

	/* We will get this response
		{
	    "data": {
	        "Get": {
	            "className	": [
	                {
	                    "answer": "DNA",
	                    "category": "SCIENCE",
	                    "question": "In 1953 Watson & Crick built a model of the molecular structure of this, the gene-carrying substance"
	                },
	                {
	                    "answer": "Liver",
	                    "category": "SCIENCE",
	                    "question": "This organ removes excess glucose from the blood & stores it as glycogen"
	                }
	            ]
	        }
	    }
		}
	*/
	if getResp, ok := resp.Data["Get"].(map[string]interface{}); ok {
		if output, ok = getResp[className].([]interface{}); !ok {
			err = errors.New("can't parse weaviate response")
		}
	}
	return
}

func (W *WeaviateVectorStore) Search(ctx context.Context, className string, query string) (output []interface{}, err error) {
	vectorQuery, err := W.embeddingModel.EmbedQuery(query)
	if err != nil {
		return
	}

	output, err = W.SearchVector(ctx, className, vectorQuery)

	return
}

func (W *WeaviateVectorStore) AddText(ctx context.Context, className string, input string) (err error) {
	_, err = W.AddDocuments(ctx, className, []string{input})
	return
}

func (W *WeaviateVectorStore) AddDocuments(ctx context.Context, className string, documents []string) (batchErr []error, err error) {
	err = W.createClassIfNotExist(ctx, className)
	if err != nil {
		return
	}

	objVectors, err := W.embeddingModel.EmbedDocuments(documents)
	if err != nil {
		return
	}
	objs := documentsToObject(className, documents, objVectors)
	batchResp, err := W.client.Batch().ObjectsBatcher().WithObjects(objs...).Do(ctx)
	if err != nil {
		return
	}
	for _, res := range batchResp {
		if res.Result.Errors != nil {
			batchErr = append(batchErr, errors.New(res.Result.Errors.Error[0].Message))
		}
	}

	return
}

func documentsToObject(className string, documents []string, vectors [][]float32) (objs []*models.Object) {
	for idx, doc := range documents {
		objs = append(objs, &models.Object{
			Class: className,
			Properties: map[string]any{
				"text": doc,
			},
			Vector: vectors[idx],
		})
	}
	return
}

func (W *WeaviateVectorStore) createClassIfNotExist(ctx context.Context, className string) (err error) {
	classExist, err := W.isClassExist(ctx, className)
	if !classExist {
		// create classHere
		err = W.createClass(ctx, className)
		if err != nil {
			return
		}
	}

	return
}

// createClass with default schema
func (W *WeaviateVectorStore) createClass(ctx context.Context, className string) (err error) {
	classSchema := &models.Class{
		Class: className,
		Properties: []*models.Property{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
		},
	}
	err = W.client.Schema().ClassCreator().WithClass(classSchema).Do(ctx)

	return
}

// isClassExist check existance of a class
func (W *WeaviateVectorStore) isClassExist(ctx context.Context, className string) (exist bool, err error) {
	if val, ok := W.existClass[className]; ok {
		return val, nil
	}
	exist, err = W.client.Schema().ClassExistenceChecker().WithClassName(className).Do(ctx)
	if err != nil {
		return
	}
	W.existClass[className] = exist

	return
}

// DeleteIndex will delete a class
func (W *WeaviateVectorStore) DeleteIndex(ctx context.Context, className string) (err error) {
	err = W.client.Schema().ClassDeleter().WithClassName(className).Do(ctx)
	return
}