package service

import (
	"context"

	stuPb "github.com/1055373165/ggcache/api/studentpb"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/internal/bussiness/student/ecode"
)

type StudentSrv struct {
	stuPb.UnimplementedStudentServiceServer
}

func NewStudentSrv() (*StudentSrv, error) {
	if err := dao.InitDB(); err != nil {
		return nil, err
	}
	return &StudentSrv{}, nil
}

func (s *StudentSrv) StudentShow(ctx context.Context, req *stuPb.StudentRequest) (*stuPb.StudentDetailResponse, error) {
	resp := &stuPb.StudentDetailResponse{}
	resp.Code = ecode.SUCCESS

	stu, err := dao.NewStudentDao(ctx).ShowStudentInfo(req)
	if err != nil {
		resp.Code = ecode.ERROR
		return nil, err
	}
	resp.StudentDetail = &stuPb.StudentResponse{
		Name:  stu.Name,
		Score: float32(stu.Score),
	}

	return resp, nil
}

func (s *StudentSrv) StudentCreate(ctx context.Context, req *stuPb.StudentRequest) (*stuPb.StudentCommonResonse, error) {
	resp := &stuPb.StudentCommonResonse{}
	resp.Code = ecode.SUCCESS

	err := dao.NewStudentDao(ctx).CreateStudent(req)
	if err != nil {
		resp.Code = ecode.ERROR
		return nil, err
	}

	resp.Message = ecode.GetMsg(int(resp.Code))
	return resp, nil
}

func (s *StudentSrv) StudentDelete(ctx context.Context, req *stuPb.StudentRequest) (*stuPb.StudentCommonResonse, error) {
	resp := &stuPb.StudentCommonResonse{}
	resp.Code = ecode.SUCCESS
	return resp, nil
}

func (s *StudentSrv) StudentUpdate(ctx context.Context, req *stuPb.StudentRequest) (*stuPb.StudentCommonResonse, error) {
	resp := &stuPb.StudentCommonResonse{}
	resp.Code = ecode.SUCCESS
	return resp, nil
}

func (s *StudentSrv) StudentLogout(ctx context.Context, request *stuPb.StudentRequest) (resp *stuPb.StudentCommonResonse, err error) {
	return
}
