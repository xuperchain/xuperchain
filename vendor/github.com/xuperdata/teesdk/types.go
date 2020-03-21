package teesdk 

type FuncCaller struct {
    Method string  `json:"method"`
    Args string     `json:"args"`
    Svn uint32   `json:"svn"`
    Address string `json:"address"`
}

type KMSCaller struct {
    Method string `json:"method"`// init
    Kds string   `json:"kds"`
    Svn uint32   `json:"svn"`
}
