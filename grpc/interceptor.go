package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/budhip/common/auth"
	svcerr "github.com/budhip/common/error"
	rc "github.com/budhip/common/remoteconfig"
	"github.com/budhip/common/tls"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	errorCode    = "error_code"
	errorMessage = "error_message"
)

func errorMetadata(serviceError svcerr.ServiceError) metadata.MD {
	data := make(map[string]string)
	data[errorCode] = serviceError.Code
	data[errorMessage] = serviceError.Message
	if len(serviceError.Attributes) > 0 {
		for k, v := range serviceError.Attributes {
			data[k] = v
		}
	}
	return metadata.New(data)
}

func grpcError(serviceError svcerr.ServiceError) error {
	pbError := Error{
		Code:       serviceError.Code,
		Message:    serviceError.Message,
		Attributes: serviceError.Attributes,
	}
	grpcError := status.New(serviceError.Status, serviceError.Message)
	errWithDetails, err := grpcError.WithDetails(&pbError)
	if err != nil {
		return grpcError.Err()
	}
	return errWithDetails.Err()
}

// UnaryErrorInterceptor returns a new unary server interceptor that added error detail for service error.
func UnaryErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			var serviceError svcerr.ServiceError
			if ok := errors.As(err, &serviceError); ok {
				md := errorMetadata(serviceError)
				if err := grpc.SetTrailer(ctx, md); err != nil {
					log.Print(err)
				}

				return nil, grpcError(serviceError)
			}

			return nil, err
		}

		return resp, nil
	}
}

// StreamErrorInterceptor returns a new streaming server interceptor that added error detail for service error.
func StreamErrorInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if err != nil {
			var serviceError svcerr.ServiceError
			if ok := errors.As(err, &serviceError); ok {
				return grpcError(serviceError)
			}

			return err
		}

		return nil
	}
}

// UnaryAuthInterceptor returns a new unary server interceptor that extract user info from token.
func UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		ctx = auth.WithUserInfoContext(ctx)
		return handler(ctx, req)
	}
}

func UnaryRecoveryInterceptor() grpc.UnaryServerInterceptor {
	handler := func(p interface{}) (err error) {
		log.Print("panic", p)
		return status.Error(codes.Internal, "server panic")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(handler),
	}

	return recovery.UnaryServerInterceptor(opts...)
}

func UnaryMaintenanceInterceptor(firebaseClientEmail, firebaseClientPrivatekey string,
	projectID string, baseURL string, environment string,
	serviceMap map[string]string, billpaymentReq map[int]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		getToken, err := WithGoogleServiceAccount(firebaseClientEmail, firebaseClientPrivatekey)

		if err != nil {
			log.Printf("token error : %v", err)
		}

		if getToken == nil {
			return handler(ctx, req)
		}

		for key, val := range serviceMap {
			if info.FullMethod == val {
				var RemoteConfigURL = baseURL + "/v1/projects/" + projectID + "/remoteConfig"
				var client = &http.Client{}
				reqt, err := http.NewRequest("GET", RemoteConfigURL, nil)
				if err != nil {
					log.Printf("Error: %v\n", err)
				}

				// Set Authorization Header
				log.Printf("firebase request started...")
				reqt.Header.Set("Authorization", "Bearer "+getToken.AccessToken)
				resp, err := client.Do(reqt)

				if err != nil {
					log.Printf("Error: %v\n", err)
				}
				log.Printf("firebase connected...")

				defer resp.Body.Close()

				// if resp.Status is not 200
				if resp.StatusCode == http.StatusOK {

					errMaintenance := WithFirebasePayloads(resp.Body, key, environment, req, billpaymentReq)

					if errMaintenance != nil {
						// return maintenance
						return ctx, errMaintenance
					}

				}

			}
		}
		return handler(ctx, req)
	}
}

func StreamRecoveryInterceptor() grpc.StreamServerInterceptor {
	handler := func(p interface{}) (err error) {
		log.Print("panic", p)
		return status.Error(codes.Internal, "server panic")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(handler),
	}

	return recovery.StreamServerInterceptor(opts...)
}

func recoveryInterceptor() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	handler := func(p interface{}) (err error) {
		log.Print("panic", p)
		return status.Error(codes.Internal, "server panic")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(handler),
	}

	return recovery.UnaryServerInterceptor(opts...), recovery.StreamServerInterceptor(opts...)
}

// WithRecovery return gRPC server options with recovery handler
func WithRecovery() []grpc.ServerOption {
	unary, stream := recoveryInterceptor()
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(unary),
		grpc.StreamInterceptor(stream),
	}
	return serverOptions
}

// WithValidation returns gRPC server options with request validator
func WithValidation() []grpc.ServerOption {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(validator.UnaryServerInterceptor()),
		grpc.StreamInterceptor(validator.StreamServerInterceptor()),
	}
	return serverOptions
}

// WithErrorDetails returns gRPC server options with request validator
func WithErrorDetails() []grpc.ServerOption {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(UnaryErrorInterceptor()),
		grpc.StreamInterceptor(StreamErrorInterceptor()),
	}
	return serverOptions
}

// WithSecure returns gRPC server option with SSL credentials
func WithSecure(ca, cert, key []byte, mutual bool) grpc.ServerOption {
	tlsCfg := tls.WithCertificate(ca, cert, key, mutual)
	return grpc.Creds(credentials.NewTLS(tlsCfg))
}

// WithDefault returns default gRPC server option with validation, recovery, auth and error interceptor
func WithDefault() []grpc.ServerOption {
	unaryRecovery, streamRecovery := recoveryInterceptor()
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			unaryRecovery,
			validator.UnaryServerInterceptor(),
			UnaryAuthInterceptor(),
			UnaryErrorInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			streamRecovery,
			validator.StreamServerInterceptor(),
			StreamErrorInterceptor(),
		)}
	return serverOptions
}

func WithGoogleServiceAccount(firebaseClientEmail, firebaseClientPrivatekey string) (*oauth2.Token, error) {

	config := &jwt.Config{
		Email:      firebaseClientEmail,
		PrivateKey: []byte(firebaseClientPrivatekey),
		Scopes: []string{
			"https://www.googleapis.com/auth/firebase.remoteconfig",
		},
		TokenURL: google.JWTTokenURL,
	}
	token, err := config.TokenSource(context.Background()).Token()
	if err != nil {
		return nil, err
	}
	return token, nil
}

func WithFirebasePayloads(resp io.ReadCloser, key string, environment string, req interface{},
	billpaymentReq map[int]string) error {
	var bytes []byte
	// Read response body
	bodyBytes, err := ioutil.ReadAll(resp)
	if err != nil {
		log.Print(err)
	}
	// Convert Response Body to String
	bodyString := string(bodyBytes)

	rawIn := json.RawMessage(bodyString)

	bytes, err = rawIn.MarshalJSON()
	if err != nil {
		log.Print(err)
	}

	var payload rc.FbRemoteConfig
	if err = json.Unmarshal(bytes, &payload); err != nil {
		log.Print(err)
	}

	VirgoFeatureFlag := payload.ParameterGroups.VirgoFeatureFlag.Parameters

	brequest := convertIntrefaceToStruct(req)
	if brequest.ProductType > 0 && brequest.ProductType < 5 {
		err = checkMaintenanceStatusForBillPaymentService(brequest, billpaymentReq, environment, VirgoFeatureFlag)
		return err
	}

	if brequest.ProductType == 0 {
		err = checkMaintenanceStatusForTransactionService(environment, key, VirgoFeatureFlag)
		return err
	}

	return nil
}

func checkMaintenanceStatusForTransactionService(environment string, key string,
	virgoFeatureFlag map[string]rc.VirgoFeatureFlagServices) error {
	for featureName, val := range virgoFeatureFlag {
		if featureName == key {
			for env, v := range val.ConditionalValues {
				if env == environment && strings.ToUpper(v.Value) == rc.Maintenance {
					return svcerr.ServiceError{
						Code:    rc.Maintenance,
						Status:  codes.Internal,
						Message: rc.Maintenance,
					}
				}
			}
			return nil
		}
	}
	return nil
}

func checkMaintenanceStatusForBillPaymentService(brequest rc.BillPaymentRequest,
	billpaymentReq map[int]string, environment string,
	virgoFeatureFlag map[string]rc.VirgoFeatureFlagServices) error {
	for featureName, val := range virgoFeatureFlag {
		productType := brequest.ProductType
		if item, ok := billpaymentReq[productType]; ok {
			if featureName == item {
				for env, v := range val.ConditionalValues {
					if env == environment && strings.ToUpper(v.Value) == rc.Maintenance {
						return svcerr.ServiceError{
							Code:    rc.Maintenance,
							Status:  codes.Internal,
							Message: rc.Maintenance,
						}
					}
				}
				return nil
			}
		}
	}
	return nil
}

func convertIntrefaceToStruct(event interface{}) rc.BillPaymentRequest {
	c := rc.BillPaymentRequest{}
	err := mapstructure.Decode(event, &c)
	if err != nil {
		log.Println(err)
	}
	return c
}