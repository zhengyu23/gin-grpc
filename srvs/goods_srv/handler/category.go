package handler

// // CreateCategory 接口实现错误！
import (
	"context"
	"encoding/json"
	"xrUncle/srvs/goods_srv/global"
	"xrUncle/srvs/goods_srv/model"
	"xrUncle/srvs/goods_srv/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetAllCategorysList 获取商品分类列表, 返回json文件
func (s *GoodsServer) GetAllCategorysList(context.Context, *emptypb.Empty) (*proto.CategoryListResponse, error) {
	var categorys []model.Category

	//global.DB.Preload("SubCategory").Find(&categorys)

	// 预加载 Preload
	// "SubCategory.SubCategory": 因为商品分类高度最高3层，除了子分类，还需加载子分类的子分类
	// 若高度为4,5,6...层，则需继续 Preload
	global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categorys)

	//for _, category := range categorys {
	//	fmt.Println(category.Name)
	//}

	// 返回数据为json格式
	b, _ := json.Marshal(&categorys)
	return &proto.CategoryListResponse{JsonData: string(b)}, nil
}

// GetSubCategory 获取子分类
func (s *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	var category model.Category
	var categoryListResponse proto.SubCategoryListResponse
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}

	categoryListResponse.Info = &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}
	// 预加载子分类
	var subCategorys []model.Category
	var subCategoryResponse []*proto.CategoryInfoResponse
	preloads := "SubCategory"
	if category.Level == 1 {
		preloads = "SubCategory.SubCategory"
	}
	global.DB.Where(&model.Category{ParentCategoryID: req.Id}).Preload(preloads).Find(&subCategorys)
	for _, subCategory := range subCategorys {
		subCategoryResponse = append(subCategoryResponse, &proto.CategoryInfoResponse{
			Id:             subCategory.ID,
			Name:           subCategory.Name,
			ParentCategory: subCategory.ParentCategoryID,
			Level:          subCategory.Level,
			IsTab:          subCategory.IsTab,
		})
	}
	categoryListResponse.SubCategorys = subCategoryResponse

	return &categoryListResponse, nil
}

// CreateCategory 新建分类信息
func (s *GoodsServer) CreateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	category := model.Category{}

	category.Name = req.Name
	category.Level = req.Level
	if req.Level != 1 {
		// 查询父目录是否存在
		var categoryParent model.Category
		categoryParent.ParentCategoryID = req.ParentCategory
		if result := global.DB.Preload("ParentCategory").First(&categoryParent, req.ParentCategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "父类目不存在")
		}
		if categoryParent.Level >= req.Level {
			return nil, status.Errorf(codes.InvalidArgument, "子父类目不可同级；或子类目level不可小于父类目level")
		}
	}
	category.ParentCategoryID = req.ParentCategory
	category.IsTab = req.IsTab
	global.DB.Create(&category)
	return &proto.CategoryInfoResponse{
		Id:             int32(category.ID),
		Name:           category.Name,
		Level:          category.Level,
		ParentCategory: category.ParentCategoryID,
		IsTab:          category.IsTab,
	}, nil
}

// DeleteCategory 删除分类
func (s *GoodsServer) DeleteCategory(ctx context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Category{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	return &emptypb.Empty{}, nil
}

// UpdateCategory 修改分类信息
func (s *GoodsServer) UpdateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	category := &model.Category{}
	if result := global.DB.First(category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}

	// 通过判断是否为默认值，判断商品数据是否正常
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.ParentCategory != 0 {
		category.ParentCategoryID = req.ParentCategory
	}
	if req.Level != 0 {
		category.Level = req.Level
	}
	if req.IsTab {
		category.IsTab = req.IsTab
	}

	global.DB.Save(&category)
	return &emptypb.Empty{}, nil
}
