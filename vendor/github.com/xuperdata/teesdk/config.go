package teesdk

type TEEConfig struct {
	Svn		uint32 `yaml:"svn"`
        Enable          bool `yaml:"enable"`
        TMSPort         int32 `yaml:"tmsport"`
        TDFSPort        int32 `yaml:"tdfsport"`
        Uid             string `yaml:"uid"`
        Token           string `yaml:"token"`
        Auditors        []*TEEAuditors `yaml:"auditors"`
}

type TEEAuditors struct {
        PublicDer         string `yaml:"publicder"`
        Sign              string `yaml:"sign"`
        EnclaveInfoConfig string `yaml:"enclaveinfoconfig"`
}
