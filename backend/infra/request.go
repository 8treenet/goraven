package infra

import (
	"io/ioutil"
	"strings"

	"encoding/json"
	"reflect"

	"github.com/8treenet/freedom"
	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *Request {
			return &Request{}
		})
		initiator.InjectController(func(ctx freedom.Context) (com *Request) {
			initiator.FetchInfra(ctx, &com)
			return
		})
	})
}

type Request struct {
	freedom.Infra
}

func (req *Request) BeginRequest(worker freedom.Worker) {
	req.Infra.BeginRequest(worker)
}

func (req *Request) ReadJSON(obj interface{}, validates ...bool) error {
	rawData, err := ioutil.ReadAll(req.Worker().IrisContext().Request().Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(rawData, obj); err != nil {
		return err
	}
	if len(validates) == 0 || !validates[0] {
		return nil
	}

	return req.validate(obj)
}

func (req *Request) ReadQuery(obj interface{}, validates ...bool) error {
	if err := req.Worker().IrisContext().ReadQuery(obj); err != nil {
		return err
	}
	if len(validates) == 0 || !validates[0] {
		return nil
	}
	return validate.Struct(obj)
}

func (req *Request) ReadForm(obj interface{}, validates ...bool) error {
	if err := req.Worker().IrisContext().ReadForm(obj); err != nil {
		return err
	}
	if len(validates) == 0 || !validates[0] {
		return nil
	}
	return req.validate(obj)
}

func (req *Request) validate(obj interface{}) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			if err := validate.Struct(val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}
	return validate.Struct(obj)
}

func (req *Request) GetUserId() string {
	v := req.Worker().Store().Get(UserIdStoreKey)
	if v == nil {
		return ""
	}
	if userId, ok := v.(string); ok {
		return userId
	}
	return ""
}

func GetUserName(worker freedom.Worker) string {
	v := worker.Store().Get(UserNameStoreKey)
	if v == nil {
		return ""
	}
	if userId, ok := v.(string); ok {
		return userId
	}
	return ""
}

func IsAdmin(worker freedom.Worker) bool {
	v := worker.Store().Get(UserRoleStoreKey)
	if role, ok := v.(uint8); ok && role == 1 {
		return true
	}
	return false
}

func (req *Request) GetUserName() string {
	return GetUserName(req.Worker())
}

func (req *Request) GetRole() uint8 {
	v := req.Worker().Store().Get(UserRoleStoreKey)
	if v == nil {
		return 0
	}
	if role, ok := v.(uint8); ok {
		return role
	}
	return 0
}

func (req *Request) GetToken() string {
	ctx := req.Worker().IrisContext()
	token := ctx.GetHeader("X-Access-Token")
	if token != "" {
		return token
	}

	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	token = ctx.URLParamDefault("access_token", "")
	return token
}
