package inventory_test

import (
	"ans-spareparts-api/internal/features/inventory"
	mocks "ans-spareparts-api/internal/mock"
	"ans-spareparts-api/pkg/apperror"
	"ans-spareparts-api/pkg/testutil/fixtures"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestSuite struct {
	Service       inventory.Service
	MockInventory *mocks.InventoryRepository
	Ctx           context.Context
}

func NewTestSuite() *TestSuite {
	return &TestSuite{}
}

func (ts *TestSuite) SetupTest(t *testing.T) {
	ts.MockInventory = mocks.NewMockInventoryRepository()
	ts.Service = inventory.NewService(ts.MockInventory)
	ts.Ctx = context.Background()

	t.Cleanup(func() {
		ts.MockInventory.AssertExpectations(t)
	})
}
func TestInventoryService_GetByID(t *testing.T) {
	mockinv := fixtures.ValidInventory()
	tests := []struct {
		name      string
		input     uint
		setup     func(*TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *inventory.Item)
	}{
		{
			name:  "getbyid_successfull",
			input: 1,
			setup: func(ts *TestSuite) {
				ts.MockInventory.On("GetByID", ts.Ctx, uint(1)).Return(mockinv, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, uint(1), i.ID)
			},
		},
		{
			name:  "getbyid_error_notfound",
			input: 1,
			setup: func(ts *TestSuite) {
				ts.MockInventory.On("GetByID", ts.Ctx, uint(1)).Return(nil, apperror.ErrNotFound).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name:  "getbyid_dberror",
			input: 999,
			setup: func(ts *TestSuite) {
				ts.MockInventory.On("GetByID", ts.Ctx, uint(999)).Return(nil, apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)
			inv, err := ts.Service.GetInventoryByID(ts.Ctx, test.input)

			test.assertErr(t, err)
			test.validate(t, inv)
		})
	}
}

func TestInventoryService_GetByProductID(t *testing.T) {
	mockinv := fixtures.ValidInventory()

	tests := []struct {
		name      string
		input     uint
		setup     func(*TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *inventory.Item)
	}{
		{
			name:  "getbyproductid_successfull",
			input: uint(1),
			setup: func(ts *TestSuite) {
				ts.MockInventory.On("GetByProductID", ts.Ctx, uint(1)).Return(mockinv, nil)
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.NotNil(t, i)
				assert.Equal(t, uint(1), i.ProductID)
			},
		},
		{
			name:  "getbyproductid_error_notfound",
			input: uint(999),
			setup: func(ts *TestSuite) {
				ts.MockInventory.On("GetByProductID", ts.Ctx, uint(999)).Return(nil, apperror.ErrNotFound)
			},
			assertErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, apperror.ErrNotFound, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name:  "getbyproductid_dberror",
			input: uint(1),
			setup: func(ts *TestSuite) {
				ts.MockInventory.On("GetByProductID", ts.Ctx, uint(1)).Return(nil, apperror.ErrInternalServer)
			},

			assertErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)

			inv, err := ts.Service.GetInventoryByProductID(ts.Ctx, test.input)

			test.assertErr(t, err)

			if test.validate != nil {
				test.validate(t, inv)
			}
		})
	}
}

func TestInventoryService_List(t *testing.T) {
	mockinv := fixtures.ValidListInventory()

	tests := []struct {
		name      string
		input     inventory.ListQuery
		setup     func(*TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *inventory.ListOutput)
	}{
		{
			name: "TestListService_successfull",
			input: inventory.ListQuery{
				Limit:  10,
				Offset: 0,
			},
			setup: func(ts *TestSuite) {
				var input = inventory.ListQuery{
					Limit:  10,
					Offset: 0,
				}
				total := int64(len(mockinv))
				ts.MockInventory.On("List", ts.Ctx, input).Return(mockinv, total, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.Nil(t, err)
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, lo *inventory.ListOutput) {
				mockinv := fixtures.ValidListInventory()
				var items []*inventory.Item
				for _, inv := range mockinv {
					items = append(items, &inventory.Item{
						ID:        inv.ID,
						ProductID: inv.ProductID,
						Quantity:  inv.Quantity,
					})
				}
				listoutput := &inventory.ListOutput{
					Items: items,
					Total: int64(len(items)),
				}
				assert.NotNil(t, lo)
				assert.Equal(t, listoutput.Total, lo.Total)
				assert.Equal(t, listoutput.Items[0].ID, lo.Items[0].ID)
			},
		},
		{
			name: "TestListService_dberror",
			input: inventory.ListQuery{
				Limit:  10,
				Offset: 0,
			},
			setup: func(ts *TestSuite) {
				query := inventory.ListQuery{
					Limit:  10,
					Offset: 0,
				}
				ts.MockInventory.On("List", ts.Ctx, query).Return(nil, int64(0), apperror.ErrInternalServer).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, lo *inventory.ListOutput) {
				assert.Nil(t, lo)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)

			list, err := ts.Service.List(ts.Ctx, test.input)

			test.assertErr(t, err)
			test.validate(t, list)
		})
	}
}

func TestInventoryService_UpdateQuantity(t *testing.T) {
	mockinv := fixtures.ValidInventory()

	tests := []struct {
		name      string
		id        uint
		input     inventory.UpdateQuantityInput
		setup     func(*TestSuite)
		assertErr func(*testing.T, error)
		validate  func(*testing.T, *inventory.Item)
	}{
		{
			name: "updatequantity_positivevalue_successfull",
			id:   uint(1),
			input: inventory.UpdateQuantityInput{
				ProductID: 1,
				Quantity:  1,
			},
			setup: func(ts *TestSuite) {
				mockinv := fixtures.ValidInventory()
				ts.MockInventory.On("GetByID", ts.Ctx, uint(1)).Return(mockinv, nil).Once()
				ts.MockInventory.On("UpdateQuantity", ts.Ctx, uint(1), int(1)).Return(nil)
			},
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				mockinv := fixtures.ValidInventory()
				mockinv.Quantity++
				assert.NotNil(t, i)
				assert.Equal(t, uint(1), i.ID)
				assert.Equal(t, 2, i.Quantity)
			},
		},
		{
			name: "updatequantity_negativevalue_successfull",
			id:   uint(1),
			input: inventory.UpdateQuantityInput{
				ProductID: 1,
				Quantity:  -1,
			},
			setup: func(ts *TestSuite) {
				mockinv := fixtures.ValidInventory()
				ts.MockInventory.On("GetByID", ts.Ctx, uint(1)).Return(mockinv, nil).Once()
				ts.MockInventory.On("UpdateQuantity", ts.Ctx, uint(1), int(-1)).Return(nil)
			},
			assertErr: func(t *testing.T, err error) {
				assert.Nil(t, err)
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				mockinv.Quantity--
				assert.NotNil(t, i)
				assert.Equal(t, mockinv.ID, i.ID)
				assert.Equal(t, mockinv.Quantity, i.Quantity)
			},
		},
		{
			name: "updatequantity_negativevalue_error_outofstock",
			id:   1,
			input: inventory.UpdateQuantityInput{
				ProductID: 1,
				Quantity:  -2,
			},
			setup: func(ts *TestSuite) {
				mockinv := fixtures.ValidInventory()
				ts.MockInventory.On("GetByID", ts.Ctx, uint(1)).Return(mockinv, nil).Once()
			},
			assertErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.ErrorIs(t, apperror.ErrInsufficientStock, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.Nil(t, i)
			},
		},
		{
			name: "updatequantity_dberror",
			id:   uint(1),
			input: inventory.UpdateQuantityInput{
				ProductID: 1,
				Quantity:  1,
			},
			setup: func(ts *TestSuite) {
				mockinv := fixtures.ValidInventory()
				ts.MockInventory.On("GetByID", ts.Ctx, uint(1)).Return(mockinv, nil).Once()
				ts.MockInventory.On("UpdateQuantity", ts.Ctx, uint(1), int(1)).Return(apperror.ErrInternalServer)
			},
			assertErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.ErrorIs(t, apperror.ErrInternalServer, err)
			},
			validate: func(t *testing.T, i *inventory.Item) {
				assert.Nil(t, i)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := NewTestSuite()
			ts.SetupTest(t)

			test.setup(ts)

			item, err := ts.Service.UpdateQuantity(ts.Ctx, test.id, test.input)

			test.assertErr(t, err)
			test.validate(t, item)
		})
	}
}
