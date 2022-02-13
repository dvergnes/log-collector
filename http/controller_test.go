package http_test

import (
	"encoding/json"
	"io"
	gohttp "net/http"
	"net/http/httptest"
	"strings"

	"github.com/dvergnes/log-collector/http"

	"github.com/julienschmidt/httprouter"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

var _ = Describe("Controller", func() {

	const logFolder = "/var/log"

	Describe("validateFileParameter", func() {
		When("file name is valid", func() {
			It("should not return an error", func() {
				Expect(http.ValidateFileParameter("foo.log")).Should(Succeed())
			})
		})

		When("file name is invalid", func() {
			DescribeTable("should return an error", func(file string, msg string) {
				err := http.ValidateFileParameter(file)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(msg))
			},
				Entry("empty file name", "", "file name must not be empty"),
				Entry("file is a relative path", "../../etc/master.passwd", "file name must not contain any relative or absolute path reference"),
			)
		})

	})

	Describe("parseLimit", func() {
		const max = 100
		When("limit is valid", func() {
			It("should return the limit as a uint", func() {
				l, err := http.ParseLimit(max, "30")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(l).Should(Equal(uint(30)))
			})
		})

		When("limit is empty", func() {
			It("should return 0", func() {
				l, err := http.ParseLimit(max, "")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(l).Should(BeZero())
			})
		})

		When("limit is invalid", func() {
			DescribeTable("should return an error", func(limit string, msg string) {
				_, err := http.ParseLimit(max, limit)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(msg))
			},
				Entry("limit is not an integer", "I am not an integer", "limit is not a valid integer"),
				Entry("limit is negative", "-5", "limit must be strictly positive"),
				Entry("limit is too big", "200", "limit must be equal or less than 100"),
			)

		})

	})

	Describe("checkFile", func() {
		var fs afero.Fs
		BeforeEach(func() {
			fs = afero.NewMemMapFs()
			Expect(fs.MkdirAll(logFolder, 0755)).Should(Succeed())
			_, err := fs.Create("/var/log/foo.log")
			Expect(err).ShouldNot(HaveOccurred())
		})
		When("file is valid", func() {
			It("should not return an error", func() {
				Expect(http.CheckFile(fs, "/var/log/foo.log")).Should(Succeed())
			})
		})

		When("file is invalid", func() {
			DescribeTable("should return an error", func(path string, msg string) {
				err := http.CheckFile(fs, path)
				Expect(err).Should(MatchError(msg))
			},
				Entry("file does not exist", "/var/log/not_found", "file /var/log/not_found was not found"),
				Entry("file is a directory", logFolder, "file /var/log is a directory"),
			)
		})

	})

	Describe("logHandler", func() {
		var (
			fs afero.Fs
			h  httprouter.Handle
		)
		BeforeEach(func() {
			fs = afero.NewMemMapFs()
			Expect(fs.MkdirAll(logFolder, 0755)).Should(Succeed())
			h = http.LogHandler(fs, http.Config{
				BufferSize: 1024,
				LogFolder:  logFolder,
				MaxEvents:  2,
			}, zap.NewNop())
			afero.WriteFile(fs, logFolder+"/foo.log", []byte(
				`128.84.140.215 - 0000001 [05/Oct/2020:10:32:51 -0800] "HEAD /web_assets/flash/runner/Leaderboard1_v04.swf HTTP/1.1" 200 - "http://sourceforge.net/forum/forum.php?forum_id=544686" "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1) ; .NET CLR 1.1.4322; InfoPath.2)" "128.84.140.215.6087629394390023"
190.134.145.226 - 0000002 [05/Oct/2020:10:32:51 -0800] "GET /web_assets/flash/runner/Leaderboard1_v04.swf HTTP/1.1" 200 185077 "http://sourceforge.net/project/showfiles.php?group_id=32993&package_id=25487&release_id=273294" "Mozilla/5.0 (Windows; U; Windows NT 5.0; en-US; rv:1.8.1.11) Gecko/20071127 Firefox/2.0.0.11" "190.134.145.226.6087629394390021"
159.226.247.165 - 0000003 [05/Oct/2020:10:32:51 -0800] "GET /web_assets/flash/runner/Letterboard_A_1.swf HTTP/1.1" 200 14140 "http://sourceforge.net/forum/?group_id=110672" "Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.8.1.9) Gecko/20071025 Firefox/2.0.0.9" "159.226.247.165.6087629394390022"
42.123.97.195 - 0000004 [05/Oct/2020:10:32:52 -0800] "GET /themes/acme_com/img/skins/white/logos/downloads/windows-logo.jpg HTTP/1.1" 200 3852 "http://sourceforge.net/project/showfiles.php?group_id=171676&package_id=196263&release_id=431372" "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; Media Center PC 3.0; .NET CLR 1.0.3705; .NET CLR 1.1.4322)" "42.123.97.195.6087629394390027"
240.54.187.93 - 0000005 [05/Oct/2020:10:32:52 -0800] "GET /web_assets/flash/runner/Imagine_Leaderboard.swf HTTP/1.1" 302 - "http://www.acme.com/" "Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; .NET CLR 3.0.04506)" "240.54.187.93.6087629394390025"
`), 0755)
		})
		When("parameters are invalid", func() {
			DescribeTable("should return an error response", func(params []string, msg string) {
				query := strings.Join(params, "&")
				req := httptest.NewRequest("GET", "http://localhost:8888/log?"+query, nil)
				w := httptest.NewRecorder()

				h(w, req, httprouter.Params{})

				resp := w.Result()
				Expect(resp.Header.Get("Content-Type")).Should(Equal("application/json"))
				body, _ := io.ReadAll(resp.Body)
				err := http.ErrorResponse{}
				Expect(json.Unmarshal(body, &err)).Should(Succeed())
				Expect(resp.StatusCode).Should(Equal(gohttp.StatusBadRequest))
				Expect(err.Code).Should(Equal("invalid.parameter"))
				Expect(err.Details).Should(Equal(msg))
			},
				Entry("file is empty", nil, "file name must not be empty"),
				Entry("limit is invalid", []string{"file=foo.log", "limit=-1"}, "limit must be strictly positive"),
				Entry("file is a directory", []string{"file=.", "limit=1"}, "file /var/log is a directory"),
			)

		})

		When("parameters are valid", func() {
			DescribeTable("should return the events", func(params []string, events []string) {
				query := strings.Join(params, "&")
				req := httptest.NewRequest("GET", "http://localhost:8888/log?file=foo.log&"+query, nil)
				w := httptest.NewRecorder()

				h(w, req, httprouter.Params{})

				resp := w.Result()
				Expect(resp.Header.Get("Content-Type")).Should(Equal("application/json"))
				body, _ := io.ReadAll(resp.Body)
				lr := http.LogResponse{}
				Expect(json.Unmarshal(body, &lr)).Should(Succeed())
				Expect(resp.StatusCode).Should(Equal(gohttp.StatusOK))
				Expect(lr.File).Should(Equal("/var/log/foo.log"))
				Expect(lr.Events).Should(HaveLen(len(events)))
				Expect(lr.Events).Should(ContainElements(events))
			},
				Entry("limit is unspecified", nil, []string{
				`240.54.187.93 - 0000005 [05/Oct/2020:10:32:52 -0800] "GET /web_assets/flash/runner/Imagine_Leaderboard.swf HTTP/1.1" 302 - "http://www.acme.com/" "Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; .NET CLR 3.0.04506)" "240.54.187.93.6087629394390025"`,
				`42.123.97.195 - 0000004 [05/Oct/2020:10:32:52 -0800] "GET /themes/acme_com/img/skins/white/logos/downloads/windows-logo.jpg HTTP/1.1" 200 3852 "http://sourceforge.net/project/showfiles.php?group_id=171676&package_id=196263&release_id=431372" "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; Media Center PC 3.0; .NET CLR 1.0.3705; .NET CLR 1.1.4322)" "42.123.97.195.6087629394390027"`,
				}),
				Entry("filter is applied", []string{"limit=2", "filter=HEAD"}, []string{
				`128.84.140.215 - 0000001 [05/Oct/2020:10:32:51 -0800] "HEAD /web_assets/flash/runner/Leaderboard1_v04.swf HTTP/1.1" 200 - "http://sourceforge.net/forum/forum.php?forum_id=544686" "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1) ; .NET CLR 1.1.4322; InfoPath.2)" "128.84.140.215.6087629394390023"`}),
			)
		})
	})

})
