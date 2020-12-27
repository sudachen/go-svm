package svm

func EncodeAppTemplate(version int, name string, code []byte, dataLayout DataLayout) ([]byte, error) {
	return cSvmEncodeAppTemplate(version, name, code, dataLayout)
}

func EncodeSpawnApp(version int, templateAddr Address, name string, ctorName string, calldata []byte) ([]byte, error) {
	return cSvmEncodeSpawnApp(version, templateAddr, name, ctorName, calldata)
}

func EncodeAppTx(version int, appAddr Address, funcName string, calldata []byte) ([]byte, error) {
	return cSvmEncodeAppTx(version, appAddr, funcName, calldata)
}
