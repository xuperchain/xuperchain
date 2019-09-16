package utils

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"

	"github.com/xuperchain/xuperunion/crypto/hash"

	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/pb"
)

func encodeChainedBftPhaseMessage(msg *pb.ChainedBftPhaseMessage) ([]byte, error) {
	var msgBuf bytes.Buffer
	encoder := json.NewEncoder(&msgBuf)
	if err := encoder.Encode(msg.Type); err != nil {
		return nil, err
	}
	if err := encoder.Encode(msg.ViewNumber); err != nil {
		return nil, err
	}
	if err := encoder.Encode(msg.ProposalQC); err != nil {
		return nil, err
	}
	if err := encoder.Encode(msg.JustifyQC); err != nil {
		return nil, err
	}
	return msgBuf.Bytes(), nil
}

// MakePhaseMsgDigest make ChainedBftPhaseMessage Digest
func MakePhaseMsgDigest(msg *pb.ChainedBftPhaseMessage) ([]byte, error) {
	msgEncoder, err := encodeChainedBftPhaseMessage(msg)
	if err != nil {
		return nil, err
	}
	msg.MsgDigest = hash.DoubleSha256(msgEncoder)
	return hash.DoubleSha256(msgEncoder), nil
}

// VerifyPhaseMsgDigest verify ChainedBftPhaseMessage Digest
func VerifyPhaseMsgDigest(msg *pb.ChainedBftPhaseMessage) ([]byte, bool, error) {
	if msg.GetMsgDigest() == nil {
		return nil, false, errors.New("VerifyMsgDigest error for msgDigest is nil")
	}
	msgDigest, err := MakePhaseMsgDigest(msg)
	if err != nil {
		return nil, false, err
	}
	return msgDigest, bytes.Equal(msg.GetMsgDigest(), msgDigest), nil
}

// MakePhaseMsgSign make ChainedBftPhaseMessage sign
func MakePhaseMsgSign(cryptoClient crypto_base.CryptoClient, privateKey *ecdsa.PrivateKey,
	msg *pb.ChainedBftPhaseMessage) (*pb.ChainedBftPhaseMessage, error) {
	msgDigest, err := MakePhaseMsgDigest(msg)
	if err != nil {
		return nil, err
	}
	msg.MsgDigest = msgDigest
	sign, err := cryptoClient.SignECDSA(privateKey, msgDigest)
	if err != nil {
		return nil, err
	}
	msg.Signature.Sign = sign
	return msg, nil
}

// VerifyPhaseMsgSign verify ChainedBftPhaseMessage Sign
func VerifyPhaseMsgSign(cryptoClient crypto_base.CryptoClient, msg *pb.ChainedBftPhaseMessage) (bool, error) {
	msgDigest, ok, err := VerifyPhaseMsgDigest(msg)
	if !ok || err != nil {
		return false, err
	}

	ak, err := cryptoClient.GetEcdsaPublicKeyFromJSON([]byte(msg.GetSignature().GetPublicKey()))
	if err != nil {
		return false, err
	}

	addr, err := cryptoClient.GetAddressFromPublicKey(ak)
	if err != nil {
		return false, err
	}

	if addr != msg.GetSignature().GetAddress() {
		return false, errors.New("VerifyPhaseMsgSign error, addr not match pk")
	}
	return cryptoClient.VerifyECDSA(ak, msg.GetSignature().GetSign(), msgDigest)
}

// MakeVoteMsgSign make ChainedBftVoteMessage sign
func MakeVoteMsgSign(cryptoClient crypto_base.CryptoClient, privateKey *ecdsa.PrivateKey,
	sig *pb.SignInfo, msg []byte) (*pb.SignInfo, error) {
	sign, err := cryptoClient.SignECDSA(privateKey, msg)
	if err != nil {
		return nil, err
	}
	sig.Sign = sign
	return sig, nil
}

// VerifyVoteMsgSign verify ChainedBftVoteMessage sign
func VerifyVoteMsgSign(cryptoClient crypto_base.CryptoClient, sig *pb.SignInfo, msg []byte) (bool, error) {
	ak, err := cryptoClient.GetEcdsaPublicKeyFromJSON([]byte(sig.GetPublicKey()))
	if err != nil {
		return false, err
	}

	addr, err := cryptoClient.GetAddressFromPublicKey(ak)
	if err != nil {
		return false, err
	}
	if addr != sig.GetAddress() {
		return false, errors.New("VerifyVoteMsgSign error, addr not match pk")
	}
	return cryptoClient.VerifyECDSA(ak, sig.GetSign(), msg)
}
