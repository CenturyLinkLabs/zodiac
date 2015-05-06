package discovery

type Cluster interface {
	Endpoints() []Endpoint
}

type Endpoint struct {
	URL string
}

type HardcodedCluster []Endpoint

func (c HardcodedCluster) Endpoints() []Endpoint {
	return c
}

//type FileCluster struct {
//  Path      string
//  endpoints []Endpoint
//}

//func NewFileCluster(p string) FileCluster {
//}

//func (f *FileCluster) Endpoints() []Endpoint {
//  return nil
//}
