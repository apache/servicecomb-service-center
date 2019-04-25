package config

import "testing"

func TestDefaultVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.Verification()
}

func TestBindAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.BindAddr = "abcde"
	conf.Verification()
}

func TestRPCAddrAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.RPCAddr = "abcde"
	conf.Verification()
}

func TestLocalBindAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.BindAddr = "127.0.0.1"
	conf.Verification()
}

func TestLocalRPCAddrVerification(t *testing.T) {
	conf := DefaultConfig()
	conf.RPCAddr = "127.0.0.1"
	conf.Verification()
}
