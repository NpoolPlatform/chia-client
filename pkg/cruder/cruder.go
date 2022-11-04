package cruder

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"

	npool "github.com/NpoolPlatform/message/npool"
	"google.golang.org/protobuf/types/known/structpb"
)

type Any interface{}

const (
	EQ   = "eq"
	NEQ  = "neq"
	GT   = "gt"
	GTE  = "gte"
	LT   = "lt"
	LTE  = "lte"
	IN   = "in"
	LIKE = "like"
)

type Cond struct {
	Op  string
	Val Any
}

type Fields map[string]Any

func NewFields() Fields {
	return map[string]Any{}
}

func (f Fields) WithField(field string, val Any) Fields {
	f[field] = val
	return f
}

type FilterFields map[string]*structpb.Value

func NewFilterFields() FilterFields {
	return map[string]*structpb.Value{}
}

func (f FilterFields) WithField(field string, val *structpb.Value) FilterFields {
	f[field] = val
	return f
}

type Conds map[string]*Cond

func NewConds() Conds {
	return map[string]*Cond{}
}

func (c Conds) WithCond(field string, op string, val Any) Conds {
	c[field] = &Cond{
		Op:  op,
		Val: val,
	}
	return c
}

type FilterConds map[string]*npool.FilterCond

func NewFilterConds() FilterConds {
	return map[string]*npool.FilterCond{}
}

func (c FilterConds) WithCond(filed string, op string, val *structpb.Value) FilterConds {
	c[filed] = &npool.FilterCond{
		Op:  op,
		Val: val,
	}
	return c
}

type Cruder interface {
	Create(ctx context.Context, in Any) (Any, error)
	CreateBulk(ctx context.Context, in []Any) ([]Any, error)

	Update(ctx context.Context, in Any) (Any, error)
	UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]Any) (Any, error)
	AddFields(ctx context.Context, id uuid.UUID, fields map[string]Any) (Any, error)

	Row(ctx context.Context, id uuid.UUID) (Any, error)
	Rows(ctx context.Context, conds map[string]Cond, offset, limit uint32) ([]Any, error)

	Exist(ctx context.Context, id uuid.UUID) (bool, error)
	ExistConds(ctx context.Context, conds map[string]Cond) (bool, error)
	Count(ctx context.Context, conds map[string]Cond) (uint32, error)
	// TODO: MAP | REDUCE | FILTER | SUM

	Delete(ctx context.Context, id uuid.UUID) (Any, error)
}

func AnyTypeInt32(v Any) (int32, error) {
	switch v.(type) {
	case int32:
		return v.(int32), nil
	case int:
		return int32(v.(int)), nil
	case float64:
		return int32(v.(float64)), nil
	}
	return -1, fmt.Errorf("invalid value type: %v (int32)", reflect.TypeOf(v))
}

func AnyTypeUint32(v Any) (uint32, error) {
	switch v.(type) {
	case uint32:
		return v.(uint32), nil
	case int32:
		return uint32(v.(int32)), nil
	case int:
		return uint32(v.(int)), nil
	case float64:
		return uint32(v.(float64)), nil
	}
	return 0, fmt.Errorf("invalid value type: %v (uint32)", reflect.TypeOf(v))
}

func AnyTypeFloat64(v Any) (float64, error) {
	switch v.(type) {
	case uint32:
		return float64(v.(uint32)), nil
	case int32:
		return float64((v.(int32))), nil
	case int:
		return float64((v.(int))), nil
	case float64:
		return v.(float64), nil
	}
	return 0, fmt.Errorf("invalid value type: %v (float64)", reflect.TypeOf(v))
}

func AnyTypeUUID(v Any) (uuid.UUID, error) {
	if _, ok := v.(uuid.UUID); ok {
		return v.(uuid.UUID), nil
	}
	if _, ok := v.(string); !ok {
		return uuid.UUID{}, fmt.Errorf("invalid value type: %v (uuid)", reflect.TypeOf(v))
	}
	return uuid.Parse(v.(string))
}

func AnyTypeUUIDs(v Any) ([]uuid.UUID, error) {
	if _, ok := v.([]uuid.UUID); ok {
		return v.([]uuid.UUID), nil
	}
	if _, ok := v.([]string); !ok {
		return []uuid.UUID{}, fmt.Errorf("invalid value type: %v ([]uuid)", reflect.TypeOf(v))
	}
	uuids := []uuid.UUID{}
	for _, val := range v.([]string) {
		uuidP, err := AnyTypeUUID(val)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, uuidP)
	}
	return uuids, nil
}

func AnyTypeString(v Any) (string, error) {
	if _, ok := v.(string); !ok {
		return "", fmt.Errorf("invalid value type: %v (string)", reflect.TypeOf(v))
	}
	return v.(string), nil
}
