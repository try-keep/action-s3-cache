# S3 Cache for GitHub Actions

[![Release](https://github.com/try-keep/action-s3-cache/actions/workflows/build_and_publish.yml/badge.svg?branch=main)](https://github.com/try-keep/action-s3-cache/actions/workflows/build_and_publish.yml)

[![Test](https://github.com/try-keep/action-s3-cache/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/try-keep/action-s3-cache/actions/workflows/test.yml)

### Archiving artifacts

```yml
- name: Save cache
  uses: try-keep/action-s3-cache@v1
  with:
    action: put
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1 # Or whatever region your bucket was created
    bucket: your-bucket
    s3-class: ONEZONE_IA # It's STANDARD by default. It can be either STANDARD,
    # REDUCED_REDUDANCY, ONEZONE_IA, INTELLIGENT_TIERING, GLACIER, DEEP_ARCHIVE or STANDARD_IA.
    key: ${{ runner.os }}-yarn-${{ hashFiles('yarn.lock') }}
    default-key: ${{ runner.os }}-yarn
    artifacts: |
      node_modules/*
```

### Retrieving artifacts

```yml
- name: Retrieve cache
  uses: try-keep/action-s3-cache@v1
  with:
    action: get
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1
    bucket: your-bucket
    key: ${{ runner.os }}-yarn-${{ hashFiles('yarn.lock') }}
    default-key: ${{ runner.os }}-yarn
```

### Clear cache

```yml
- name: Clear cache
  uses: try-keep/action-s3-cache@v1
  with:
    action: delete
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1
    bucket: your-bucket
    key: ${{ runner.os }}-yarn-${{ hashFiles('yarn.lock') }}
    default-key: ${{ runner.os }}-yarn
```

## Example

The following example shows a simple pipeline using S3 Cache GitHub Action:

```yml
- name: Checkout
  uses: actions/checkout@v2

- name: Retrieve cache
  uses: try-keep/action-s3-cache@v1
  with:
    action: get
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1
    bucket: your-bucket
    key: ${{ runner.os }}-yarn-${{ hashFiles('yarn.lock') }}
    default-key: ${{ runner.os }}-yarn

- name: Install dependencies
  run: yarn

- name: Save cache
  uses: try-keep/action-s3-cache@v1
  with:
    action: put
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1
    bucket: your-bucket
    s3-class: STANDARD_IA
    key: ${{ runner.os }}-yarn-${{ hashFiles('yarn.lock') }}
    default-key: ${{ runner.os }}-yarn
    artifacts: |
      node_modules/*
```
