package v1

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var _ = Describe("Nginx Webhook", func() {

	// Mutationのテスト
	Context("Mutation", func() {
		It("Should mutate a Nginx", func() {
			mutateTest(filepath.Join("testdata", "mutate", "before.yaml"), filepath.Join("testdata", "mutate", "after.yaml"))
		})
	})

	// Validationのテスト
	Context("Validation", func() {
		It("Should create a valid Nginx", func() {
			validateTest(filepath.Join("testdata", "validate", "valid.yaml"), true)
		})
		It("Should not create a invalid Nginx", func() {
			validateTest(filepath.Join("testdata", "validate", "invalid.yaml"), false)
		})
	})
})

// Mutationテスト用の関数
func mutateTest(before string, after string) {
	ctx := context.Background()

	// ファイルをByte型配列として読み込む
	y, err := os.ReadFile(before)
	Expect(err).NotTo(HaveOccurred())

	// 読み取ったYAMLの内容を&Nginx型に変換する
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(y), 4096)
	beforeNginx := &Nginx{}
	err = d.Decode(beforeNginx)
	Expect(err).NotTo(HaveOccurred())

	// Nginxリソースを作成
	err = k8sClient.Create(ctx, beforeNginx)
	Expect(err).NotTo(HaveOccurred())

	// 作成したNginxリソースを取得
	ret := &Nginx{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: beforeNginx.GetName(), Namespace: beforeNginx.GetNamespace()}, ret)
	Expect(err).NotTo(HaveOccurred())

	y, err = os.ReadFile(after)
	Expect(err).NotTo(HaveOccurred())

	d = yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(y), 4096)
	afterNginx := &Nginx{}
	err = d.Decode(afterNginx)
	Expect(err).NotTo(HaveOccurred())

	// 実際に作成されたNginxとafterのannotationsを比較する
	Expect(ret.ObjectMeta.Annotations).Should(Equal(afterNginx.ObjectMeta.Annotations))
}

// Validationテスト用の関数
//
//	第1引数: 適用するYAML
//	第2引数: 適用時のvalidation期待値
func validateTest(file string, valid bool) {
	ctx := context.Background()

	// ファイルをByte型配列として読み込む
	y, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())

	// 読み取ったYAMLの内容を&Nginx型に変換する
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(y), 4096)
	nginx := &Nginx{}
	err = d.Decode(nginx)
	Expect(err).NotTo(HaveOccurred())

	// Nginxリソースを作成
	err = k8sClient.Create(ctx, nginx)

	if valid {
		Expect(err).NotTo(HaveOccurred(), "Nginx: %v", nginx)
	} else {
		// errが値を持っていること
		Expect(err).To(HaveOccurred(), "Nginx: %v", nginx)

		// metav1.Status構造体を初期化
		// Validation Webhookは失敗するとerrors.StatusError型のエラーを返す
		statusErr := &apierrors.StatusError{}
		// metav1.Statusに該当するフィールドがerrに含まれていることを確認
		Expect(errors.As(err, &statusErr)).To(BeTrue())

		// エラーメッセージの期待値を設定
		expected := "Invalid value"

		// エラーレスポンスに含まれるmetav1.Status.Messageに期待値が含まれることを確認
		Expect(statusErr.ErrStatus.Message).To(ContainSubstring(expected))

	}

}
