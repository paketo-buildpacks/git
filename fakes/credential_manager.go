package fakes

import "sync"

type CredentialManager struct {
	SetupCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir   string
			PlatformPath string
		}
		Returns struct {
			Err error
		}
		Stub func(string, string) error
	}
}

func (f *CredentialManager) Setup(param1 string, param2 string) error {
	f.SetupCall.Lock()
	defer f.SetupCall.Unlock()
	f.SetupCall.CallCount++
	f.SetupCall.Receives.WorkingDir = param1
	f.SetupCall.Receives.PlatformPath = param2
	if f.SetupCall.Stub != nil {
		return f.SetupCall.Stub(param1, param2)
	}
	return f.SetupCall.Returns.Err
}
