/*
 * Copyright (c) 2022. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xuperchain/crypto/common/account"
	"github.com/xuperchain/xuperchain/models"
	"github.com/xuperchain/xuperchain/service/pb"
)

// test data path
var (
	testDataDir  = "test_data"
	validAKDir   = filepath.Join(testDataDir, "valid-ak")
	invalidAKDir = filepath.Join(testDataDir, "invalid-ak")
)

// test data
var (
	testAccount = "XC1111111111111111@xuper"

	testPublicKey  = `{"Curvname":"P-256","X":36505150171354363400464126431978257855318414556425194490762274938603757905292,"Y":79656876957602994269528255245092635964473154458596947290316223079846501380076}`
	testPrivateKey = `{"Curvname":"P-256","X":36505150171354363400464126431978257855318414556425194490762274938603757905292,"Y":79656876957602994269528255245092635964473154458596947290316223079846501380076,"D":111497060296999106528800133634901141644446751975433315540300236500052690483486}`

	mockSign = []byte("mockSign")

	secretKeySucc                             = testPrivateKey
	secretKeyGetEcdsaPrivateKeyFromJsonStrErr = "GetEcdsaPrivateKeyFromJsonStrErr"
	secretKeySignECDSAErr                     = "SignECDSAErr"

	ecdsaPrivateKeySignECDSAErr = new(ecdsa.PrivateKey)
)

func Test_listAKs(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		want    []AK
		wantErr bool
	}{
		{
			name: "normal",
			root: validAKDir,
			want: []AK{
				newAK(filepath.Join(validAKDir, "keys")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listAKs(tt.root)
			if (err != nil) != tt.wantErr {
				t.Errorf("listAKs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listAKs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAK_IsAKDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "AK dir",
			path: filepath.Join(validAKDir, "keys"),
			want: true,
		},
		{
			name: "not AK dir",
			path: filepath.Join(validAKDir, "not-ak-dir"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ak := AK{
				path: tt.path,
			}
			if got := ak.IsAKDir(); got != tt.want {
				t.Errorf("AK.IsAKDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAK_AuthRequirementFrom(t *testing.T) {
	type fields struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:   "valid AK",
			fields: fields{path: filepath.Join(validAKDir, "keys")},
			want:   "XC1111111111111111@xuper/TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY",
		},
		{
			name:    "invalid AK",
			fields:  fields{path: filepath.Join(validAKDir, "not-ak-dir")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ak := &AK{
				path: tt.fields.path,
			}
			got, err := ak.AuthRequirementFrom(testAccount)
			if (err != nil) != tt.wantErr {
				t.Errorf("AK.AuthRequirementFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AK.AuthRequirementFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAK_Info(t *testing.T) {
	type fields struct {
		path string
	}
	tests := []struct {
		name     string
		fields   fields
		wantInfo AKInfo
		wantErr  bool
	}{
		{
			name:   "valid AK",
			fields: fields{path: filepath.Join(validAKDir, "keys")},
			wantInfo: AKInfo{
				address: "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY",
				KeyPair: KeyPair{
					publicKey: testPublicKey,
					secretKey: testPrivateKey,
				},
			},
		},
		{
			name:    "AK without address",
			fields:  fields{path: filepath.Join(invalidAKDir, "no-address")},
			wantErr: true,
		},
		{
			name:    "AK without key",
			fields:  fields{path: filepath.Join(invalidAKDir, "no-public-key")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ak := &AK{
				path: tt.fields.path,
			}
			gotInfo, err := ak.Info()
			if (err != nil) != tt.wantErr {
				t.Errorf("AK.Info() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("AK.Info() = %v, \n want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}

func TestAK_keyPair(t *testing.T) {
	type fields struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		want    KeyPair
		wantErr bool
	}{
		{
			name:   "valid key pair",
			fields: fields{path: filepath.Join(validAKDir, "keys")},
			want: KeyPair{
				publicKey: testPublicKey,
				secretKey: testPrivateKey,
			},
		},
		{
			name:    "no public key",
			fields:  fields{path: filepath.Join(invalidAKDir, "no-public-key")},
			wantErr: true,
		},
		{
			name:    "no secret key",
			fields:  fields{path: filepath.Join(invalidAKDir, "no-private-key")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ak := &AK{
				path: tt.fields.path,
			}
			got, err := ak.keyPair()
			if (err != nil) != tt.wantErr {
				t.Errorf("AK.keyPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AK.keyPair() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAK_SignTx(t *testing.T) {
	type fields struct {
		path string
	}
	type args struct {
		tx *pb.Transaction
	}
	tx := new(pb.Transaction)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.SignatureInfo
		wantErr bool
	}{

		{
			name:   "valid AK",
			fields: fields{path: filepath.Join(validAKDir, "keys")},
			args:   args{tx: tx},
			want: &pb.SignatureInfo{
				PublicKey: testPublicKey,
				Sign:      mockSign,
			},
		},
		{
			name:    "AK without key",
			args:    args{tx: tx},
			fields:  fields{path: filepath.Join(invalidAKDir, "no-public-key")},
			wantErr: true,
		},
		{
			name:    "key pair sign fail",
			args:    args{tx: nil},
			fields:  fields{path: filepath.Join(validAKDir, "keys")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ak := &AK{
				path: tt.fields.path,
			}
			got, err := ak.SignTx(tt.args.tx, mockCryptoClient{})
			if (err != nil) != tt.wantErr {
				t.Errorf("AK.SignTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AK.SignTx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAKInfo_SignUtxo(t *testing.T) {
	type fields struct {
		KeyPair KeyPair
	}
	tests := []struct {
		name    string
		fields  fields
		want    pb.SignatureInfo
		wantErr bool
	}{
		{
			name:    "succeed",
			fields:  fields{KeyPair: KeyPair{secretKey: secretKeySucc}},
			want:    pb.SignatureInfo{
				PublicKey: testPublicKey,
				Sign:      mockSign,
			},
		},
		{
			name:    "sign fail",
			fields:  fields{KeyPair: KeyPair{secretKey: secretKeySignECDSAErr}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &AKInfo{
				address: testAccount,
				KeyPair: KeyPair{
					publicKey: testPublicKey,
					secretKey: tt.fields.KeyPair.secretKey,
				},
			}
			got, err := i.SignUtxo("xuper", big.NewInt(0), mockCryptoClient{})
			if (err != nil) != tt.wantErr {
				t.Errorf("AKInfo.SignUtxo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AKInfo.SignUtxo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: mock common.ComputeTxSign after unit test covers
func TestKeyPair_SignTx(t *testing.T) {
	type fields struct {
		secretKey string
	}
	type args struct {
		tx *pb.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.SignatureInfo
		wantErr bool
	}{
		{
			name:   "tx is normal",
			fields: fields{secretKey: secretKeySucc},
			args:   args{tx: new(pb.Transaction)},
			want: &pb.SignatureInfo{
				PublicKey: testPublicKey,
				Sign:      mockSign,
			},
		},
		{
			name:    "tx is nil",
			fields:  fields{secretKey: secretKeySucc},
			args:    args{tx: nil},
			wantErr: true,
		},
		{
			name:    "GetEcdsaPrivateKeyFromJsonStr() fail",
			fields:  fields{secretKey: secretKeyGetEcdsaPrivateKeyFromJsonStrErr},
			args:    args{tx: new(pb.Transaction)},
			wantErr: true,
		},
		{
			name:    "SignECDSA() fail",
			fields:  fields{secretKey: secretKeySignECDSAErr},
			args:    args{tx: new(pb.Transaction)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &KeyPair{
				publicKey: testPublicKey,
				secretKey: tt.fields.secretKey,
			}
			got, err := p.SignTx(tt.args.tx, mockCryptoClient{})
			if (err != nil) != tt.wantErr {
				t.Errorf("KeyPair.SignTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyPair.SignTx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyPair_SignUtxo(t *testing.T) {
	type fields struct {
		secretKey string
	}
	tests := []struct {
		name    string
		fields  fields
		want    pb.SignatureInfo
		wantErr bool
	}{
		{
			name:   "succeed",
			fields: fields{secretKey: secretKeySucc},
			want: pb.SignatureInfo{
				PublicKey: testPublicKey,
				Sign:      mockSign,
			},
		},
		{
			name:    "GetEcdsaPrivateKeyFromJsonStr() fail",
			fields:  fields{secretKey: secretKeyGetEcdsaPrivateKeyFromJsonStrErr},
			wantErr: true,
		},
		{
			name:    "SignECDSA() fail",
			fields:  fields{secretKey: secretKeySignECDSAErr},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &KeyPair{
				publicKey: testPublicKey,
				secretKey: tt.fields.secretKey,
			}
			utxo := models.NewLockedUtxo("xuper", testAccount, big.NewInt(1))
			got, err := p.SignUtxo(utxo, mockCryptoClient{})
			if (err != nil) != tt.wantErr {
				t.Errorf("KeyPair.SignUtxo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyPair.SignUtxo() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockCryptoClient struct {
}

func (m mockCryptoClient) GenerateKeyBySeed(_ []byte) (*ecdsa.PrivateKey, error) {
	panic("implement me")
}

func (m mockCryptoClient) SignECDSA(key *ecdsa.PrivateKey, _ []byte) (signature []byte, err error) {
	if key == ecdsaPrivateKeySignECDSAErr {
		return nil, errors.New("SignECDSAErr")
	}
	return mockSign, nil
}

func (m mockCryptoClient) VerifyECDSA(_ *ecdsa.PublicKey, _, _ []byte) (valid bool, err error) {
	panic("implement me")
}

func (m mockCryptoClient) VerifyXuperSignature(_ []*ecdsa.PublicKey, _ []byte, _ []byte) (valid bool, err error) {
	panic("implement me")
}

func (m mockCryptoClient) EncryptByEcdsaKey(_ *ecdsa.PublicKey, _ []byte) (cypherText []byte, err error) {
	panic("implement me")
}

func (m mockCryptoClient) DecryptByEcdsaKey(_ *ecdsa.PrivateKey, _ []byte) (msg []byte, err error) {
	panic("implement me")
}

func (m mockCryptoClient) GetAddressFromPublicKey(_ *ecdsa.PublicKey) (string, error) {
	panic("implement me")
}

func (m mockCryptoClient) CheckAddressFormat(_ string) (bool, uint8) {
	panic("implement me")
}

func (m mockCryptoClient) VerifyAddressUsingPublicKey(_ string, _ *ecdsa.PublicKey) (bool, uint8) {
	panic("implement me")
}

func (m mockCryptoClient) GetBinaryEcdsaPrivateKeyFromFile(_ string, _ string) ([]byte, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetEcdsaPrivateKeyFromFile(_ string) (*ecdsa.PrivateKey, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetEcdsaPrivateKeyFromFileByPassword(_ string, _ string) (*ecdsa.PrivateKey, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetEcdsaPublicKeyFromFile(_ string) (*ecdsa.PublicKey, error) {
	panic("implement me")
}

func (m mockCryptoClient) GenerateEntropy(_ int) ([]byte, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetEcdsaPrivateKeyFromJsonStr(secretKey string) (*ecdsa.PrivateKey, error) {
	switch secretKey {
	case secretKeyGetEcdsaPrivateKeyFromJsonStrErr:
		return nil, errors.New("GetEcdsaPrivateKeyFromJsonStrErr")
	case secretKeySignECDSAErr:
		return ecdsaPrivateKeySignECDSAErr, nil
	default:
		return nil, nil
	}
}

func (m mockCryptoClient) GetEcdsaPublicKeyFromJsonStr(_ string) (*ecdsa.PublicKey, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetEcdsaPrivateKeyJsonFormatStr(_ *ecdsa.PrivateKey) (string, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetEcdsaPublicKeyJsonFormatStr(_ *ecdsa.PrivateKey) (string, error) {
	panic("implement me")
}

func (m mockCryptoClient) ExportNewAccount(_ string) error {
	panic("implement me")
}

func (m mockCryptoClient) CreateNewAccountWithMnemonic(_ int, _ uint8) (*account.ECDSAAccount, error) {
	panic("implement me")
}

func (m mockCryptoClient) ExportNewAccountWithMnemonic(_ string, _ int, _ uint8) error {
	panic("implement me")
}

func (m mockCryptoClient) RetrieveAccountByMnemonic(_ string, _ int) (*account.ECDSAAccount, error) {
	panic("implement me")
}

func (m mockCryptoClient) RetrieveAccountByMnemonicAndSavePrivKey(_ string, _ int, _ string, _ string) (*account.ECDSAInfo, error) {
	panic("implement me")
}

func (m mockCryptoClient) EncryptAccount(_ *account.ECDSAAccount, _ string) (*account.ECDSAAccountToCloud, error) {
	panic("implement me")
}

func (m mockCryptoClient) GenerateMnemonic(_ []byte, _ int) (string, error) {
	panic("implement me")
}

func (m mockCryptoClient) GenerateSeedWithErrorChecking(_ string, _ string, _ int, _ int) ([]byte, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetRandom32Bytes() ([]byte, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetRiUsingRandomBytes(_ *ecdsa.PublicKey, _ []byte) []byte {
	panic("implement me")
}

func (m mockCryptoClient) GetRUsingAllRi(_ *ecdsa.PublicKey, _ [][]byte) []byte {
	panic("implement me")
}

func (m mockCryptoClient) GetSharedPublicKeyForPublicKeys(_ []*ecdsa.PublicKey) ([]byte, error) {
	panic("implement me")
}

func (m mockCryptoClient) GetSiUsingKCRM(_ *ecdsa.PrivateKey, _ []byte, _ []byte, _ []byte, _ []byte) []byte {
	panic("implement me")
}

func (m mockCryptoClient) GetSUsingAllSi(_ [][]byte) []byte {
	panic("implement me")
}

func (m mockCryptoClient) GenerateMultiSignSignature(_ []byte, _ []byte) ([]byte, error) {
	panic("implement me")
}

func (m mockCryptoClient) VerifyMultiSig(_ []*ecdsa.PublicKey, _, _ []byte) (bool, error) {
	panic("implement me")
}

func (m mockCryptoClient) MultiSign(_ []*ecdsa.PrivateKey, _ []byte) ([]byte, error) {
	panic("implement me")
}
