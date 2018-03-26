package kubernetestracker_test

import (
	. "github.com/dgruber/drmaa2os/pkg/jobtracker/kubernetestracker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"

	"time"
)

var _ = Describe("KubernetesTracker", func() {

	Context("Basic interface test", func() {
		var kt jobtracker.JobTracker
		var jt drmaa2interface.JobTemplate

		BeforeEach(func() {
			jt = drmaa2interface.JobTemplate{
				//JobName:       "name1",
				RemoteCommand: "/bin/sh",
				JobCategory:   "golang:latest",
				Args:          []string{"-c", "sleep 0"},
			}
			var err error
			kt, err = New()
			Ω(err).Should(BeNil())
		})

		WhenK8sIsAvailableIt("should be possible to AddJob()", func() {
			jobid, err := kt.AddJob(jt)
			Ω(err).Should(BeNil())
			Ω(jobid).ShouldNot(Equal(""))
		})

		WhenK8sIsAvailableIt("should be possible to AddArrayJob()", func() {
			jobid, err := kt.AddArrayJob(jt, 1, 2, 1, 0)
			Ω(err).Should(BeNil())
			Ω(jobid).ShouldNot(Equal(""))
		})

		WhenK8sIsAvailableIt("should be possible to ListJobs()", func() {
			jobids, err := kt.ListJobs()
			Ω(err).Should(BeNil())
			Ω(jobids).ShouldNot(BeNil())
		})

		WhenK8sIsAvailableIt("should be possible to ListArrayJobs()", func() {
			jobids, err := kt.ListArrayJobs("123")
			Ω(err).ShouldNot(BeNil())
			Ω(jobids).Should(BeNil())
		})

		WhenK8sIsAvailableIt("should be possible ListJobsCategories()", func() {
			cats, err := kt.ListJobCategories()
			Ω(err).Should(BeNil())
			Ω(cats).ShouldNot(BeNil())
			Ω(len(cats)).Should(BeNumerically("==", 0))
		})

	})

	Context("Unsupported interface functions", func() {
		var kt jobtracker.JobTracker

		BeforeEach(func() {
			var err error
			kt, err = New()
			Ω(err).Should(BeNil())
		})

		It("Unsupported ListJobCategories()", func() {
			_, err := kt.ListJobCategories()
			Ω(err).Should(BeNil())
		})

		It("Unsupported JobInfo()", func() {
			_, err := kt.JobInfo("jobid")
			Ω(err).Should(BeNil())
		})

		It("Unsupported DeleteJob()", func() {
			err := kt.DeleteJob("jobid")
			Ω(err).Should(BeNil())
		})

	})

	Context("Basic Kubernetes Job Workflow", func() {
		var kt jobtracker.JobTracker
		var jt drmaa2interface.JobTemplate

		BeforeEach(func() {
			jt = drmaa2interface.JobTemplate{
				//JobName:       "workfloadtestjob",
				RemoteCommand: "/bin/sh",
				JobCategory:   "golang:latest",
			}
			var err error
			kt, err = New()
			Ω(err).Should(BeNil())
		})

		WhenK8sIsAvailableIt("Should be possible to track the states of a job life-cycle", func() {
			jt.Args = []string{"-c", "sleep 1"}
			jobid, err := kt.AddJob(jt)
			Ω(err).Should(BeNil())
			Ω(jobid).ShouldNot(Equal(""))

			Eventually(func() drmaa2interface.JobState {
				return kt.JobState(jobid)
			}, time.Second*10, time.Millisecond*5).Should(Equal(drmaa2interface.Undetermined))

			Eventually(func() drmaa2interface.JobState {
				return kt.JobState(jobid)
			}, time.Second*10, time.Millisecond*50).Should(Equal(drmaa2interface.Running))

			Eventually(func() drmaa2interface.JobState {
				return kt.JobState(jobid)
			}, time.Second*30, time.Millisecond*50).Should(Equal(drmaa2interface.Done))

		})

		WhenK8sIsAvailableIt("Should be possible to terminate a job", func() {
			jt.Args = []string{"-c", "sleep 10"}
			jobid, err := kt.AddJob(jt)
			Ω(err).Should(BeNil())
			Ω(jobid).ShouldNot(Equal(""))

			Eventually(func() drmaa2interface.JobState {
				return kt.JobState(jobid)
			}, time.Second*10, time.Millisecond*20).Should(Equal(drmaa2interface.Running))

			err = kt.JobControl(jobid, "terminate")
			Ω(err).Should(BeNil())

			Eventually(func() drmaa2interface.JobState {
				return kt.JobState(jobid)
			}, time.Second*30, time.Millisecond*10).Should(Equal(drmaa2interface.Undetermined))
		})

		WhenK8sIsAvailableIt("Should be possible to wait for termination of a job", func() {
			jt.Args = []string{"-c", "sleep 10"}
			jobid, err := kt.AddJob(jt)
			Ω(err).Should(BeNil())
			Ω(jobid).ShouldNot(Equal(""))

			go func() {
				<-time.Tick(time.Millisecond * 100)
				kt.JobControl(jobid, "terminate")
			}()

			err = kt.Wait(jobid, time.Second, drmaa2interface.Failed, drmaa2interface.Undetermined)
			Ω(err).Should(BeNil())
			// TODO(DG) test time = 0
			// TODO(DG) terminate should lead to failed state not undetermined
		})

	})

})
