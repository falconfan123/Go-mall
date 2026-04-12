package server

import (
	"context"

	"github.com/falconfan123/Go-mall/services/search/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/search/internal/application/service"
	"github.com/falconfan123/Go-mall/services/search/internal/svc"
	search "github.com/falconfan123/Go-mall/services/search/pb"
)

type SearchServiceServer struct {
	search.UnimplementedSearchServiceServer
	ctx *svc.ServiceContext
	svc *service.SearchAppService
}

func NewSearchServiceServer(ctx *svc.ServiceContext) *SearchServiceServer {
	return &SearchServiceServer{
		ctx: ctx,
		svc: service.NewSearchAppService(ctx.QueryParser, ctx.ElasticsearchClient),
	}
}

func (s *SearchServiceServer) ParseQuery(ctx context.Context, req *search.ParseQueryRequest) (*search.ParseQueryResponse, error) {
	result, err := s.svc.ParseQuery(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return &search.ParseQueryResponse{
			OriginalQuery:     "",
			NormalizedQuery:   "",
			PredictedCategory: "",
			Brands:            []string{},
			ProductWords:      []string{},
			Modifiers:         []string{},
		}, nil
	}

	return &search.ParseQueryResponse{
		OriginalQuery:     result.OriginalQuery,
		NormalizedQuery:   result.NormalizedQuery,
		PredictedCategory: result.PredictedCategory,
		Brands:            result.Brands,
		ProductWords:      result.ProductWords,
		Modifiers:         result.Modifiers,
	}, nil
}

func (s *SearchServiceServer) Search(ctx context.Context, req *search.SearchRequest) (*search.SearchResponse, error) {
	dtoReq := &dto.QueryRequest{
		Query:     req.Query,
		Category:  req.Category,
		Page:      int(req.Page),
		PageSize:  int(req.PageSize),
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	result, err := s.svc.Search(ctx, dtoReq)
	if err != nil {
		return nil, err
	}

	results := make([]*search.SearchResult, len(result.Results))
	for i, r := range result.Results {
		results[i] = &search.SearchResult{
			Id:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Price:       r.Price,
			ImageUrl:    r.ImageURL,
			Category:    r.Category,
			Brand:       r.Brand,
			Score:       r.Score,
		}
	}

	return &search.SearchResponse{
		Results:    results,
		Total:      result.Total,
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
		TotalPages: int32(result.TotalPages),
	}, nil
}
