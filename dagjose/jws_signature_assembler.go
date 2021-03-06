package dagjose

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/node/mixins"
)

type jwsSignatureAssembler struct {
	signature *jwsSignature
	key       *string
	state     maState
}

var jwsSignatureMixin = mixins.MapAssembler{TypeName: "JWSSignature"}

func (j *jwsSignatureAssembler) BeginMap(sizeHint int64) (ipld.MapAssembler, error) {
	if j.state == maState_midValue && *j.key == "header" {
		j.signature.header = make(map[string]ipld.Node)
		j.state = maState_initial
		return &headerAssembler{
			header: j.signature.header,
			key:    nil,
			state:  maState_initial,
		}, nil
	}
	if j.state != maState_initial {
		panic("misuse")
	}
	return j, nil
}
func (j *jwsSignatureAssembler) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	return jwsSignatureMixin.BeginList(sizeHint)
}
func (j *jwsSignatureAssembler) AssignNull() error {
	if j.state == maState_midValue {
		switch *j.key {
		case "header":
			j.signature.header = nil
		case "protected":
			j.signature.protected = nil
		case "signature":
			j.signature.signature = nil
		default:
			panic("should never happen due to validation in map assembler")
		}
		return nil
	}
	return jwsSignatureMixin.AssignNull()
}
func (j *jwsSignatureAssembler) AssignBool(b bool) error {
	return jwsSignatureMixin.AssignBool(b)
}
func (j *jwsSignatureAssembler) AssignInt(i int64) error {
	return jwsSignatureMixin.AssignInt(i)
}
func (j *jwsSignatureAssembler) AssignFloat(f float64) error {
	return jwsSignatureMixin.AssignFloat(f)
}
func (j *jwsSignatureAssembler) AssignString(s string) error {
	if j.state == maState_midKey {
		if !isValidJWSSignatureKey(s) {
			return fmt.Errorf("%s is not a vliad JWS signature key", s)
		}
		j.key = &s
		j.state = maState_expectValue
		return nil
	}
	return jwsSignatureMixin.AssignString(s)
}
func (j *jwsSignatureAssembler) AssignBytes(b []byte) error {
	if j.state == maState_midValue {
		if *j.key == "protected" {
			j.signature.protected = b
			j.state = maState_initial
			return nil
		}
		if *j.key == "signature" {
			j.signature.signature = b
			j.state = maState_initial
			return nil
		}
		panic("should not be possible due to validation in map assembler")
	}
	return jwsSignatureMixin.AssignBytes(b)
}
func (j *jwsSignatureAssembler) AssignLink(l ipld.Link) error {
	return jwsSignatureMixin.AssignLink(l)
}
func (j *jwsSignatureAssembler) AssignNode(n ipld.Node) error {
	return fmt.Errorf("not implemented")
}
func (j *jwsSignatureAssembler) Prototype() ipld.NodePrototype {
	return basicnode.Prototype.Map
}

func (j *jwsSignatureAssembler) AssembleKey() ipld.NodeAssembler {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.state = maState_midKey
	return j
}

func (j *jwsSignatureAssembler) AssembleValue() ipld.NodeAssembler {
	if j.state != maState_expectValue {
		panic("misuse")
	}
	j.state = maState_midValue
	return j
}
func (j *jwsSignatureAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.key = &k
	j.state = maState_midValue
	return j, nil
}

func (j *jwsSignatureAssembler) KeyPrototype() ipld.NodePrototype {
	return basicnode.Prototype.String
}
func (j *jwsSignatureAssembler) ValuePrototype(k string) ipld.NodePrototype {
	return basicnode.Prototype.Any
}

func (j *jwsSignatureAssembler) Finish() error {
	if j.state != maState_initial {
		panic("misuse")
	}
	j.state = maState_finished
	return nil
}

func isValidJWSSignatureKey(key string) bool {
	return key == "protected" || key == "header" || key == "signature"
}
