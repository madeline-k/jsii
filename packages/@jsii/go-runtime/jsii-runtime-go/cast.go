package jsii

import (
	"fmt"
	"reflect"

	"github.com/aws/jsii-runtime-go/internal/kernel"
)

// UnsafeCast converts the given interface value to the desired target interface
// pointer. Panics if the from value is not a jsii proxy object, or if the to
// value is not a pointer to an interface type.
func UnsafeCast(from interface{}, into interface{}) {
	rinto := reflect.ValueOf(into)
	if rinto.Kind() != reflect.Ptr {
		panic(fmt.Errorf("Second argument to UnsafeCast must be a pointer to an interface. Received %s", rinto.Type().String()))
	}
	rinto = rinto.Elem()
	if rinto.Kind() != reflect.Interface {
		panic(fmt.Errorf("Second argument to UnsafeCast must be a pointer to an interface. Received pointer to %s", rinto.Type().String()))
	}

	rfrom := reflect.ValueOf(from)

	// If rfrom is essentially nil, set into to nil and return.
	if !rfrom.IsValid() || rfrom.IsZero() {
		null := reflect.Zero(rinto.Type())
		rinto.Set(null)
		return
	}
	// Interfaces may present as a pointer to an implementing struct, and that's fine...
	if rfrom.Kind() != reflect.Interface && rfrom.Kind() != reflect.Ptr {
		panic(fmt.Errorf("First argument to UnsafeCast must be an interface value. Received %s", rfrom.Type().String()))
	}

	// If rfrom can be directly converted to rinto, just do it.
	if rfrom.CanConvert(rinto.Type()) {
		rfrom = rfrom.Convert(rinto.Type())
		rinto.Set(rfrom)
		return
	}

	client := kernel.GetClient()
	if objID, found := client.FindObjectRef(rfrom); found {
		// Ensures the value is initialized properly. Panics if the target value is not a jsii interface type.
		client.Types().InitJsiiProxy(rinto)

		// If the target type is a behavioral interface, add it to the ObjectRef.Interfaces list.
		if fqn, found := client.Types().InterfaceFQN(rinto.Type()); found {
			objID.Interfaces = append(objID.Interfaces, fqn)
		}

		// Make the new value an alias to the old value.
		client.RegisterInstance(rinto, objID)
		return
	}

	panic(fmt.Errorf("First argument to UnsafeCast must be a jsii proxy value. Received %s", rfrom.String()))
}
