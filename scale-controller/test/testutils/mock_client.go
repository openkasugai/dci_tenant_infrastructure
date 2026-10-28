package testutils

import (
	"context"
	"errors"
	"slices"

	"go.uber.org/mock/gomock"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"scale-controller/test/testutils/mock"
)

type (
	getFunc    func(ctx context.Context, namespacedName types.NamespacedName, obj client.Object, opts ...client.GetOption) error
	listFunc   func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
	createFunc func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	deleteFunc func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
	updateFunc func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	patchFunc  func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error
	statusFunc func() client.SubResourceWriter
	schemeFunc func() *runtime.Scheme
)

type counter struct {
	count         int
	isInitialized bool
}

func (c *counter) initialize() error {
	if c.isInitialized {
		return errors.New("counter is already initialized")
	}
	c.count = 0
	c.isInitialized = true
	return nil
}

func (c *counter) increment() int {
	c.count++
	return c.count
}

type clientCounters struct {
	get    counter
	list   counter
	create counter
	delete counter
	update counter
	patch  counter
	status counter
	scheme counter
}

type MockClientHelper struct {
	realClient            client.Client
	mockClient            *mock.MockClient
	mockSubResourceWriter *mock.MockSubResourceWriter
	counters              clientCounters
}

func NewMockClientHelper(
	controller *gomock.Controller, realClient client.Client,
) *MockClientHelper {
	c := mock.NewMockClient(controller)
	sw := mock.NewMockSubResourceWriter(controller)
	sw.EXPECT().
		Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(apierrors.NewInternalError(errors.New("mocked error for Status Patch")))
	return &MockClientHelper{
		realClient:            realClient,
		mockClient:            c,
		mockSubResourceWriter: sw,
		counters:              clientCounters{},
	}
}

func (m *MockClientHelper) GetMockClient() *mock.MockClient {
	return m.mockClient
}

func (m *MockClientHelper) SetDefault() {
	m.mockClient.EXPECT().Scheme().AnyTimes().Return(m.realClient.Scheme())
	m.mockClient.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
			return m.realClient.Create(ctx, obj, opts...)
		},
	)
	m.mockClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
			return m.realClient.Update(ctx, obj, opts...)
		},
	)
	m.mockClient.EXPECT().Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			return m.realClient.Patch(ctx, obj, patch, opts...)
		},
	)
	m.mockClient.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
			return m.realClient.Delete(ctx, obj, opts...)
		},
	)
	m.mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, namespacedName types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
			return m.realClient.Get(ctx, namespacedName, obj, opts...)
		},
	)
	m.mockClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
			return m.realClient.List(ctx, list, opts...)
		},
	)
	m.mockClient.EXPECT().Status().AnyTimes().DoAndReturn(
		func() client.SubResourceWriter {
			return m.realClient.Status()
		},
	)
}

func (m *MockClientHelper) setGet(
	fn getFunc,
) error {
	if err := m.counters.get.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetGetWithFunc(
	defaultFunc, alterFunc getFunc, alterCases []int,
) error {
	fn := func(ctx context.Context, namespacedName types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
		if slices.Contains(alterCases, m.counters.get.increment()) {
			return alterFunc(ctx, namespacedName, obj, opts...)
		}
		return defaultFunc(ctx, namespacedName, obj, opts...)
	}
	return m.setGet(fn)
}

func (m *MockClientHelper) SetGet(
	errorCases []int,
) error {
	errorFunc := func(_ context.Context, _ types.NamespacedName, _ client.Object, _ ...client.GetOption) error {
		return apierrors.NewInternalError(errors.New("mocked error for Get"))
	}
	return m.SetGetWithFunc(m.realClient.Get, errorFunc, errorCases)
}

func (m *MockClientHelper) setList(
	fn listFunc,
) error {
	if err := m.counters.list.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetListWithFunc(
	defaultFunc, alterFunc listFunc, alterCases []int,
) error {
	fn := func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
		if slices.Contains(alterCases, m.counters.list.increment()) {
			return alterFunc(ctx, list, opts...)
		}
		return defaultFunc(ctx, list, opts...)
	}
	return m.setList(fn)
}

func (m *MockClientHelper) SetList(
	errorCases []int,
) error {
	errorFunc := func(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
		return apierrors.NewInternalError(errors.New("mocked error for List"))
	}
	return m.SetListWithFunc(m.realClient.List, errorFunc, errorCases)
}

func (m *MockClientHelper) setCreate(
	fn createFunc,
) error {
	if err := m.counters.create.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetCreateWithFunc(
	defaultFunc, alterFunc createFunc, alterCases []int,
) error {
	fn := func(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
		if slices.Contains(alterCases, m.counters.create.increment()) {
			return alterFunc(ctx, obj, opts...)
		}
		return defaultFunc(ctx, obj, opts...)
	}
	return m.setCreate(fn)
}

func (m *MockClientHelper) SetCreate(
	errorCases []int,
) error {
	errorFunc := func(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
		return apierrors.NewInternalError(errors.New("mocked error for Create"))
	}
	return m.SetCreateWithFunc(m.realClient.Create, errorFunc, errorCases)
}

func (m *MockClientHelper) setDelete(
	fn deleteFunc,
) error {
	if err := m.counters.delete.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		Delete(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetDeleteWithFunc(
	defaultFunc, alterFunc deleteFunc, alterCases []int,
) error {
	fn := func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
		if slices.Contains(alterCases, m.counters.delete.increment()) {
			return alterFunc(ctx, obj, opts...)
		}
		return defaultFunc(ctx, obj, opts...)
	}
	return m.setDelete(fn)
}

func (m *MockClientHelper) SetDelete(
	errorCases []int,
) error {
	errorFunc := func(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
		return apierrors.NewInternalError(errors.New("mocked error for Delete"))
	}
	return m.SetDeleteWithFunc(m.realClient.Delete, errorFunc, errorCases)
}

func (m *MockClientHelper) setUpdate(
	fn updateFunc,
) error {
	if err := m.counters.update.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		Update(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetUpdateWithFunc(
	defaultFunc, alterFunc updateFunc, alterCases []int,
) error {
	fn := func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
		if slices.Contains(alterCases, m.counters.update.increment()) {
			return alterFunc(ctx, obj, opts...)
		}
		return defaultFunc(ctx, obj, opts...)
	}
	return m.setUpdate(fn)
}

func (m *MockClientHelper) SetUpdate(
	errorCases []int,
) error {
	errorFunc := func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
		return apierrors.NewInternalError(errors.New("mocked error for Update"))
	}
	return m.SetUpdateWithFunc(m.realClient.Update, errorFunc, errorCases)
}

func (m *MockClientHelper) setPatch(
	fn patchFunc,
) error {
	if err := m.counters.patch.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetPatchWithFunc(
	defaultFunc, alterFunc patchFunc, alterCases []int,
) error {
	fn := func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
		if slices.Contains(alterCases, m.counters.patch.increment()) {
			return alterFunc(ctx, obj, patch, opts...)
		}
		return defaultFunc(ctx, obj, patch, opts...)
	}
	return m.setPatch(fn)
}

func (m *MockClientHelper) SetPatch(
	errorCases []int,
) error {
	errorFunc := func(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
		return apierrors.NewInternalError(errors.New("mocked error for Patch"))
	}
	return m.SetPatchWithFunc(m.realClient.Patch, errorFunc, errorCases)
}

func (m *MockClientHelper) setStatusPatch(
	fn statusFunc,
) error {
	if err := m.counters.status.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		Status().
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetStatusPatchWithFunc(
	defaultFunc, alterFunc statusFunc, alterCases []int,
) error {
	fn := func() client.SubResourceWriter {
		if slices.Contains(alterCases, m.counters.status.increment()) {
			return alterFunc()
		}
		return defaultFunc()
	}
	return m.setStatusPatch(fn)
}

func (m *MockClientHelper) SetStatusPatch(
	errorCases []int,
) error {
	errorFunc := func() client.SubResourceWriter {
		return m.mockSubResourceWriter
	}
	return m.SetStatusPatchWithFunc(m.realClient.Status, errorFunc, errorCases)
}

func (m *MockClientHelper) setScheme(
	fn schemeFunc,
) error {
	if err := m.counters.scheme.initialize(); err != nil {
		return err
	}
	m.mockClient.EXPECT().
		Scheme().
		AnyTimes().
		DoAndReturn(fn)
	return nil
}

func (m *MockClientHelper) SetSchemeWithFunc(
	defaultFunc, alterFunc schemeFunc, alterCases []int,
) error {
	fn := func() *runtime.Scheme {
		if slices.Contains(alterCases, m.counters.scheme.increment()) {
			return alterFunc()
		}
		return defaultFunc()
	}
	return m.setScheme(fn)
}

func (m *MockClientHelper) SetScheme(
	errorCases []int,
) error {
	return m.SetSchemeWithFunc(m.realClient.Scheme, runtime.NewScheme, errorCases)
}
