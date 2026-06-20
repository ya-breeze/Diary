package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ya-breeze/diary.be/pkg/utils"
)

var _ = Describe("ComputeTagsSourceHash", func() {
	It("is stable for identical content", func() {
		h1 := utils.ComputeTagsSourceHash("Title", "Body")
		h2 := utils.ComputeTagsSourceHash("Title", "Body")
		Expect(h1).To(Equal(h2))
		Expect(h1).NotTo(BeEmpty())
	})

	It("changes when the title changes", func() {
		Expect(utils.ComputeTagsSourceHash("A", "Body")).
			NotTo(Equal(utils.ComputeTagsSourceHash("B", "Body")))
	})

	It("changes when the body text changes", func() {
		Expect(utils.ComputeTagsSourceHash("Title", "A")).
			NotTo(Equal(utils.ComputeTagsSourceHash("Title", "B")))
	})

	It("changes when an asset is added or removed", func() {
		withImg := utils.ComputeTagsSourceHash("Title", "text ![alt](a.jpg)")
		without := utils.ComputeTagsSourceHash("Title", "text")
		Expect(withImg).NotTo(Equal(without))
	})

	It("is invariant to asset reference ordering in the body", func() {
		ab := utils.ComputeTagsSourceHash("Title", "![x](a.jpg) ![y](b.jpg)")
		ba := utils.ComputeTagsSourceHash("Title", "![y](b.jpg) ![x](a.jpg)")
		Expect(ab).To(Equal(ba))
	})
})
