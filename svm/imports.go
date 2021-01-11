package svm

import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

type Imports struct {
	_inner unsafe.Pointer

	// envs holds import functions environment objects.
	// `svm_trampoline` will get the respective environment object raw pointer directly from SVM.
	// tracking it here is needed so that it won't get GC-ed.
	envs []*functionEnvironment
}

func (imports Imports) Free() {
	cSvmImportsDestroy(imports)
}

// ImportFunction represents an SVM-runtime imported function.
type ImportFunction struct {
	// implementation represents the real function implementation written in Go.
	implementation hostFunction

	// namespace is the imported function WebAssembly namespace.
	namespace string

	// params is the WebAssembly signature of the function implementation params.
	params ValueTypes

	// returns is the WebAssembly signature of the function implementation returns.
	returns ValueTypes
}

type ImportsBuilder struct {
	imports          map[string]ImportFunction
	currentNamespace string
}

func NewImportsBuilder() ImportsBuilder {
	var imports = make(map[string]ImportFunction)
	var currentNamespace = "host"

	return ImportsBuilder{imports, currentNamespace}
}

// Namespace changes the current namespace of the next imported functions.
func (ib ImportsBuilder) Namespace(namespace string) ImportsBuilder {
	ib.currentNamespace = namespace
	return ib
}

func (ib ImportsBuilder) RegisterFunction(name string, params ValueTypes, returns ValueTypes, function hostFunction) (ImportsBuilder, error) {
	//params, returns, err := validateImport(name, impl)
	//if err != nil {
	//	return ImportsBuilder{}, err
	//}

	ib.imports[name] = ImportFunction{
		function,
		ib.currentNamespace,
		params,
		returns,
	}

	return ib, nil
}

func (ib ImportsBuilder) Build() (Imports, error) {
	imports := Imports{}
	imports.envs = make([]*functionEnvironment, 0)

	if res := cSvmImportsAlloc(&imports._inner, uint(len(ib.imports))); res != cSvmSuccess {
		return Imports{}, fmt.Errorf("failed to allocate imports")
	}

	for imprtName, imprt := range ib.imports {
		env := functionEnvironment{
			hostFunctionStoreIndex: hostFunctionStore.set(imprt.implementation),
		}
		imports.envs = append(imports.envs, &env)

		if err := cSvmImportFuncNew(
			imports,
			imprt.namespace,
			imprtName,
			unsafe.Pointer(&env),
			imprt.params,
			imprt.returns,
		); err != nil {
			return Imports{}, fmt.Errorf("failed to build import `%v`: %v", imprtName, err)
		}
	}

	return imports, nil
}

func validateImport(name string, implementation interface{}) (args ValueTypes, returns ValueTypes, err error) {
	var importType = reflect.TypeOf(implementation)

	if importType.Kind() != reflect.Func {
		err = fmt.Errorf("imported function `%s` must be a function; given `%s`", name, importType.Kind())
		return
	}

	var inputArity = importType.NumIn()

	if inputArity < 1 {
		err = fmt.Errorf("imported function `%s` must at least have one argument (for the runtime context)", name)
		return
	}

	//if importType.In(0).Kind() != reflect.UnsafePointer {
	//	err = fmt.Errorf("the runtime context of the `%s` imported function must be of kind `unsafe.Pointer`; given `%s`", name, importType.In(0).Kind())
	//	return
	//}

	inputArity--

	var outputArity = importType.NumOut()
	args = make(ValueTypes, inputArity)
	returns = make(ValueTypes, outputArity)

	for i := 0; i < inputArity; i++ {
		var importInput = importType.In(i + 1)

		switch importInput.Kind() {
		case reflect.Int32:
			args[i] = TypeI32
		case reflect.Int64:
			args[i] = TypeI64
		default:
			err = fmt.Errorf("invalid input type for the `%s` imported function; given `%s`; only accept `int32` and `int64`", name, importInput.Kind())
			return
		}
	}

	//if outputArity > 1 {
	if outputArity > 999 {
		err = fmt.Errorf("the `%s` imported function must have at most one output value", name)
		return
	} else if outputArity == 1 {
		switch importType.Out(0).Kind() {
		case reflect.Int32:
			returns[0] = TypeI32
		case reflect.Int64:
			returns[0] = TypeI64
		default:
			err = fmt.Errorf("invalid output type for the `%s` imported function; given `%s`; only accept `int32` and `int64`", name, importType.Out(0).Kind())
			return
		}
	}

	return
}
