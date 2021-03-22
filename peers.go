package SimpleCache

import pb "SimpleCache/proto"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
