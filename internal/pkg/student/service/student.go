package service

import (
	"context"

	stuPb "github.com/1055373165/ggcache/api/studentpb"
	"github.com/1055373165/ggcache/internal/pkg/student/dao"
	"github.com/1055373165/ggcache/internal/pkg/student/ecode"
)

type StudentSrv struct {
	stuPb.UnimplementedStudentServiceServer
}

func (s *StudentSrv) StudentLogin(ctx context.Context, req *stuPb.StudentRequest) (resp *stuPb.StudentDetailResponse, err error) {
	resp = new(stuPb.StudentDetailResponse)
	resp.Code = ecode.SUCCESS
	r, err := dao.NewStudentDao(ctx).ShowStudentInfo(req)
	if err != nil {
		resp.Code = ecode.ERROR
		return
	}
	resp.StudentDetail = &stuPb.StudentResponse{
		Name:  r.Name,
		Score: float32(r.Score),
	}
	return
}

func (s *StudentSrv) StudentRegister(ctx context.Context, req *stuPb.StudentRequest) (resp *stuPb.StudentCommonResonse, err error) {
	resp = new(stuPb.StudentCommonResonse)
	resp.Code = ecode.SUCCESS
	err = dao.NewStudentDao(ctx).CreateStudent(req)
	if err != nil {
		resp.Code = ecode.ERROR
		return
	}
	resp.Message = ecode.GetMsg(int(resp.Code))
	return
}

func (s *StudentSrv) StudentLogout(ctx context.Context, request *stuPb.StudentRequest) (resp *stuPb.StudentCommonResonse, err error) {
	return
}
