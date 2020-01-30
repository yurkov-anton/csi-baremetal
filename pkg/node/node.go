package node

import (
	"context"

	api "eos2git.cec.lab.emc.com/ECS/baremetal-csi-plugin.git/api/generated/v1"
	"eos2git.cec.lab.emc.com/ECS/baremetal-csi-plugin.git/pkg/sc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// store storage class name
type SCName string

type CSINodeService struct {
	VolumeManager
	scMap  map[SCName]sc.StorageClassImplementer
	NodeID string
}

// depending on SC and parameters in CreateVolumeRequest()
// here we should use different SC implementations for creating required volumes
// the same principle we can use in Controller Server or read from a CRD instance

func (s *CSINodeService) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (s *CSINodeService) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (s *CSINodeService) NodePublishVolume(ctx context.Context,
	req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	ll := logrus.WithFields(logrus.Fields{
		"component": "NodeService",
		"method":    "NodePublishVolume",
		"volumeID":  req.VolumeId,
	})
	ll.Infof("Processing request: %v", req)

	s.cacheMutex.Lock()
	ll.Info("Lock mutex")
	defer func() {
		s.cacheMutex.Unlock()
		ll.Info("Unlock mutex")
	}()

	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target Path missing in request")
	}

	v := s.getVolumeFromCache(req.VolumeId)
	if v == nil {
		return nil, status.Error(codes.NotFound, "There is no volume with appropriate VolumeID")
	}

	scImpl := s.scMap[SCName("hdd")]
	targetPath := req.TargetPath
	bdev := v.Location

	ok, _ := scImpl.CreateFileSystem(sc.XFS, bdev)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to create file system")
	}
	ok, _ = scImpl.CreateTargetPath(targetPath)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to create target path")
	}
	ok, _ = scImpl.Mount(bdev, targetPath)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to mount to target path")
	}

	v.Status = api.OperationalStatus_Operative
	ll.Infof("Successfully mount derive %s to path %s", bdev, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *CSINodeService) NodeUnpublishVolume(ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	ll := logrus.WithFields(logrus.Fields{
		"component": "NodeService",
		"method":    "NodeUnpublishVolume",
		"volumeID":  req.VolumeId,
	})
	ll.Infof("Processing request: %v", req)

	s.cacheMutex.Lock()
	ll.Info("Lock mutex")
	defer func() {
		s.cacheMutex.Unlock()
		ll.Info("Unlock mutex")
	}()

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target Path missing in request")
	}

	v := s.getVolumeFromCache(req.VolumeId)
	if ok := s.scMap["hdd"].Unmount(req.TargetPath); !ok {
		return nil, status.Error(codes.Internal, "Unable to unmount")
	}

	v.Status = api.OperationalStatus_Staging
	ll.Infof("volume was successfully unmount from %s", req.TargetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *CSINodeService) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return &csi.NodeGetVolumeStatsResponse{}, nil
}

func (s *CSINodeService) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return &csi.NodeExpandVolumeResponse{}, nil
}

func (s *CSINodeService) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func (s *CSINodeService) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	topology := csi.Topology{
		Segments: map[string]string{
			"baremetal-csi/nodeid": s.NodeID,
		},
	}

	logrus.WithFields(logrus.Fields{
		"component": "nodeService",
		"method":    "NodeGetInfo",
	}).Infof("NodeGetInfo created topology: %v", topology)

	return &csi.NodeGetInfoResponse{
		NodeId:             s.NodeID,
		AccessibleTopology: &topology,
	}, nil
}
