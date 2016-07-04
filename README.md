# DTT // dearing's template tool

Common template testing wrapped into a simple CLI.

| arg | description |
|--------|-------|
|`secret` | AWS Secret |
|`key` | AWS Key |
|`region` | AWS Region |

style
---

Accepts a list of templates to read, parse as json and then pretty-print back to the file.

| arg | description |
|--------|-------|
|`save` | boolean to save template back|

validate
---

Upload, Verify and Delete a template against S3 bucket.

| arg | description |
|--------|-------|
|`bucket` | the bucket location to use|

test
---

Execute a template pack that tests a collection of templates with defined parameters.

push
---

Push templates to a target s3 bucket.

| arg | description |
|--------|-------|
|`bucket` | s3 bucket to upload to|

pull
---

Pull down templates from an s3 bucket.

| arg | description |
|--------|-------|
|`bucket` | s3 bucket to pull from |
