package servicebroker_test

import (
	. "cf/commands/servicebroker"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callDeleteServiceBroker(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-service-broker", args)
	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteServiceBroker(ui, config, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func deleteServiceBroker(t mr.TestingT, confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) {
	serviceBroker := models.ServiceBroker{}
	serviceBroker.Name = "service-broker-to-delete"
	serviceBroker.Guid = "service-broker-to-delete-guid"

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	repo = &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	config := testconfig.NewRepositoryWithDefaults()

	ctxt := testcmd.NewContext("delete-service-broker", args)
	cmd := NewDeleteServiceBroker(ui, config, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestDeleteServiceBrokerFailsWithUsage", func() {
		ui, _, _ := deleteServiceBroker(mr.T(), "y", []string{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui, _, _ = deleteServiceBroker(mr.T(), "y", []string{"my-broker"})
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestDeleteServiceBrokerRequirements", func() {

		reqFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}

		reqFactory.LoginSuccess = false
		callDeleteServiceBroker(mr.T(), []string{"-f", "my-broker"}, reqFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory.LoginSuccess = true
		callDeleteServiceBroker(mr.T(), []string{"-f", "my-broker"}, reqFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestDeleteConfirmingWithY", func() {

		ui, _, repo := deleteServiceBroker(mr.T(), "y", []string{"service-broker-to-delete"})

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
			{"Really delete", "service-broker-to-delete"},
		})
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteConfirmingWithYes", func() {

		ui, _, repo := deleteServiceBroker(mr.T(), "Yes", []string{"service-broker-to-delete"})

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
			{"Really delete", "service-broker-to-delete"},
		})

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteWithForceOption", func() {

		serviceBroker := models.ServiceBroker{}
		serviceBroker.Name = "service-broker-to-delete"
		serviceBroker.Guid = "service-broker-to-delete-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo := &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
		ui := callDeleteServiceBroker(mr.T(), []string{"-f", "service-broker-to-delete"}, reqFactory, repo)

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))
		Expect(len(ui.Prompts)).To(Equal(0))
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteAppThatDoesNotExist", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo := &testapi.FakeServiceBrokerRepo{FindByNameNotFound: true}
		ui := callDeleteServiceBroker(mr.T(), []string{"-f", "service-broker-to-delete"}, reqFactory, repo)

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal(""))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete"},
			{"OK"},
			{"service-broker-to-delete", "does not exist"},
		})
	})
})
