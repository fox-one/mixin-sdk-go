package mixin

import (
	"crypto/sha256"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

const (
	TIPVerify                   = "TIP:VERIFY:"
	TIPAddressAdd               = "TIP:ADDRESS:ADD:"
	TIPAddressRemove            = "TIP:ADDRESS:REMOVE:"
	TIPUserDeactivate           = "TIP:USER:DEACTIVATE:"
	TIPEmergencyContactCreate   = "TIP:EMERGENCY:CONTACT:CREATE:"
	TIPEmergencyContactRead     = "TIP:EMERGENCY:CONTACT:READ:"
	TIPEmergencyContactRemove   = "TIP:EMERGENCY:CONTACT:REMOVE:"
	TIPPhoneNumberUpdate        = "TIP:PHONE:NUMBER:UPDATE:"
	TIPMultisigRequestSign      = "TIP:MULTISIG:REQUEST:SIGN:"
	TIPMultisigRequestUnlock    = "TIP:MULTISIG:REQUEST:UNLOCK:"
	TIPCollectibleRequestSign   = "TIP:COLLECTIBLE:REQUEST:SIGN:"
	TIPCollectibleRequestUnlock = "TIP:COLLECTIBLE:REQUEST:UNLOCK:"
	TIPTransferCreate           = "TIP:TRANSFER:CREATE:"
	TIPWithdrawalCreate         = "TIP:WITHDRAWAL:CREATE:"
	TIPRawTransactionCreate     = "TIP:TRANSACTION:CREATE:"
	TIPOAuthApprove             = "TIP:OAUTH:APPROVE:"
	TIPProvisioningUpdate       = "TIP:PROVISIONING:UPDATE:"
	TIPAppOwnershipTransfer     = "TIP:APP:OWNERSHIP:TRANSFER:"
	TIPSequencerRegister        = "SEQUENCER:REGISTER:"
)

func (c *Client) EncryptTipPin(key mixinnet.Key, action string, params ...string) string {
	hash := sha256.New()
	hash.Write([]byte(action))
	for _, p := range params {
		hash.Write([]byte(p))
	}

	return c.EncryptPin(key.Sign(hash.Sum(nil)).String())
}
