package server

import (
	"fmt"
)

//
// --- SendLoginPassword ---
//

// Шифрование передаваемых данных.
func layerSendLoginPasswordEncode(data TxLoginPassword, secretKey [32]byte) (eData TxLoginPassword, err error) {

	eData.TxID = data.TxID

	eData.TxFor, err = encrypt(data.TxFor, secretKey)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого txFor: <%v>", err)
	}

	eData.TxLogin, err = encrypt(data.TxLogin, secretKey)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого txLogin: <%v>", err)
	}

	eData.TxPassword, err = encrypt(data.TxPassword, secretKey)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого txPassword: <%v>", err)
	}

	eData.TxCreatedAt, err = encrypt(data.TxCreatedAt, secretKey)
	if err != nil {
		return TxLoginPassword{}, fmt.Errorf("Error: ошибка шифрования содержимого TxCreatedAt: <%v>", err)
	}

	return eData, nil
}

//
// --- SendText ---
//

// Шифрование передаваемых данных.
func layerSendTextEncode(data TxText, secretKey [32]byte) (eData TxText, err error) {

	eData.TxID = data.TxID

	eData.TxFor, err = encrypt(data.TxFor, secretKey)
	if err != nil {
		return TxText{}, fmt.Errorf("Error: ошибка шифрования содержимого txFor: <%v>", err)
	}

	eData.TxText, err = encrypt(data.TxText, secretKey)
	if err != nil {
		return TxText{}, fmt.Errorf("Error: ошибка шифрования содержимого txLogin: <%v>", err)
	}

	eData.TxCreatedAt, err = encrypt(data.TxCreatedAt, secretKey)
	if err != nil {
		return TxText{}, fmt.Errorf("Error: ошибка шифрования содержимого TxCreatedAt: <%v>", err)
	}

	return eData, nil
}
