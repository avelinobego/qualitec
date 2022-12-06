package util

import "io/ioutil"

// ExtractAssetToFile extrai um arquivo embutido no código fonte via programas como
// go-bindata e o grava em um arquivo em disco
func ExtractAssetToFile(assetName string, assetFunc func(name string) ([]byte, error), fileName string) ([]byte, error) {
	content, err := assetFunc(assetName)
	if err != nil {
		return content, err
	}
	return content, ioutil.WriteFile(fileName, content, 0)
}
