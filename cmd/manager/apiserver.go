package main

import (
	"context"
	"fmt"
	"os"
	"time"

	v1alpha1 "github.com/chechiachang/cattle-operator/pkg/apis/cattle/v1alpha1"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// === API Server ===
// FIXME move to other package after this is released:
// https://github.com/operator-framework/operator-sdk/pull/1357
// Require manager.GetClient() when NewWithClient()

var authorization string

func init() {
	authorization = os.Getenv("INTERNAL_API_AUTH_TOKEN")
}

func NewWithClient(c client.Client) *gin.Engine {

	router := gin.Default()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())

	router.GET("/ping", Ping)
	router.POST("/cattle", CreateCattle(c))

	return router
}

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// CreateCattle is a handler function which parses request and sends response
func CreateCattle(client client.Client) func(*gin.Context) {
	return func(c *gin.Context) {

		if auth := c.Request.Header.Get("Authorization"); len(auth) == 0 || auth != authorization {
			c.JSON(401, gin.H{
				"message": fmt.Sprintf("Unauthorized"),
			})
			return
		}

		// Parse & Validate Request
		var cattleCreateRequest CattleCreateRequest
		c.BindJSON(&cattleCreateRequest)
		name := cattleCreateRequest.Name
		if len(name) <= 0 {
			msg := fmt.Sprintf("Invalid token create request: %s", name)
			log.Info(msg)
			c.JSON(403, gin.H{
				"message": msg,
			})
			return
		}

		// Check if CRD already exists
		// Check if the deployment already exists, if not create a new one
		found := &v1alpha1.Cattle{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: "default"}, found)
		if err != nil && errors.IsNotFound(err) {

			log.Info("Creating a new CRD")
			crd := crdForCattle(name)
			err = client.Create(context.TODO(), crd)
			if err != nil {
				msg := fmt.Sprintf("Failed to create new Cattle CRD %s", name)
				log.Error(err, msg)
				c.JSON(500, gin.H{
					"message": msg,
				})
				return
			}

			msg := fmt.Sprintf("Successfully created new Cattle CRD %s", name)
			log.Info(msg)
			c.JSON(201, gin.H{
				"message": msg,
			})
			return

		} else if err != nil {
			msg := fmt.Sprintf("Failed to get CRD from cluster %s", name)
			log.Error(err, msg)
			c.JSON(500, gin.H{
				"message": msg,
			})
			return
		}

		msg := fmt.Sprintf("CRD already exists in cluster: %s", name)
		c.JSON(409, gin.H{
			"message": msg,
		})
		return
	}
}

// crdForCattle returns a Cattle Defination
func crdForCattle(name string) *v1alpha1.Cattle {
	return &v1alpha1.Cattle{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cattle.chechiachang.com/v1alpha1",
			Kind:       "Cattle",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1alpha1.CattleSpec{
			Name: name,
			Size: 1,
			BeefParts: []string{
				"chuck",
				"ribs",
				"plate",
			},
		}}
}

type CattleCreateRequest struct {
	Name string `json:"name"`
}
