{{/*
This is the template for the pipeline template input.
*/ -}}
# Generated using "task set-pipeline"
temp:
  repo: &repo
    type: git
    icon: github
    webhook_token: ((webhook_token))
    check_every: 24h
  repo_source: &repo_source
    uri: ((config.github.uri))
    branch: ((config.github.branch))
    private_key: ((github.private_key))
    fetch_tags: true

  docker_image_source: &docker_image_source
    repository: ((config.docker.repository))
    username: ((dockerhub.username))
    password: ((dockerhub.password))

  helm: &helm
    type: helm
    check_every: 24h
    icon: factory

  artifacts: &artifacts
    type: s3
    icon: chip
  artifacts_source: &artifacts_source
    access_key_id: ((artifacts.access_key_id))
    bucket: ((artifacts.bucket))
    region_name: ((artifacts.region_name))
    secret_access_key: ((artifacts.secret_access_key))

  github_artifact: &github_artifact
    type: github-release
    icon: hammer
    check_every: 24h
  github_artifact_source: &github_artifact_source
    access_token: ((github.access_token))

  build_task: &build_task
    image: go-build-image
    config:
      platform: linux
      inputs:
      - name: app-source
      caches:
      - path: cache/gopath
      outputs:
      - name: build
      run:
        path: sh
        args:
        - -ec
        - |
          export GOPATH=$PWD/cache/gopath

          suffix=${GOOS}-${GOARCH}
          test "$GOARCH" = "arm" && suffix="${suffix}$GOARM"

          filename=relay-${suffix}
          test "$GOOS" = "windows" && filename="${filename}.exe"
          target=$PWD/build/$filename
          dockerfile=$PWD/build/Dockerfile

          cat > $dockerfile <<EOT
          FROM scratch
          ADD $filename /relay
          ENTRYPOINT ["/relay"]
          EOT

          cd app-source
          go build -o $target -a -tags netgo -ldflags "-w" ./cmd/relay
          ls -l $target
          sha256sum $target

  image_task: &image_task
    privileged: true
    config:
      platform: linux
      image_resource:
        type: registry-image
        source:
          repository: concourse/oci-build-task
      inputs:
      - name: build
        path: .
      outputs:
      - name: image
      run:
        path: build

  alpine_task: &alpine_task
    platform: linux
    image_resource:
      type: registry-image
      source:
        repository: alpine
        tag: "3.14"

  values_task: &values_task
    config:
      <<: *alpine_task
      outputs:
      - name: values
        path: .
      run:
        path: sh
        args:
        - -ec
        - |
          echo "$VALUES_YAML" > values.yaml

  smoketest_task: &smoketest_task
    task: smoketest
    image: go-build-image
    config:
      platform: linux
      inputs:
      - name: smoketest
      caches:
      - path: cache/gopath
      run:
        path: sh
        args:
        - -ec
        - |
          export GOPATH=$PWD/cache/gopath

          cd smoketest
          go test ./test/...
          go run ./test/smoketest

resources:
- name: app-source
  <<: *repo
  source:
    <<: *repo_source
    paths:
    - cmd
    - internal
    - tools
    - go.mod
    - go.sum

- name: smoketest
  <<: *repo
  source:
    <<: *repo_source
    paths:
    - features/smoketest
    - test/smoketest
    - go.mod
    - go.sum

- name: chart-source
  <<: *repo
  source:
    <<: *repo_source
    paths:
    - helm

- name: ci-source
  <<: *repo
  source:
    <<: *repo_source
    paths:
    - ci

{{- range .platforms }}
{{- if .docker }}
- name: {{template "image" .}}
  type: registry-image
  icon: floppy
  check_every: 24h
  source:
    <<: *docker_image_source
    variant: {{template "variant" .}}
{{- end}}
{{- end}}

- name: artifacts
  <<: *artifacts
  source:
    <<: *artifacts_source
    versioned_file: ardumower-relay/rc/bundle.tgz

- name: app-artifacts
  <<: *artifacts
  source:
    <<: *artifacts_source
    versioned_file: ardumower-relay/rc/app.tgz

- name: chart-artifacts
  <<: *artifacts
  source:
    <<: *artifacts_source
    versioned_file: ardumower-relay/rc/chart.tgz

- name: chart-artifact
  type: s3
  icon: map
  source:
    access_key_id: ((artifacts.access_key_id))
    bucket: ((artifacts.bucket))
    region_name: ((artifacts.region_name))
    secret_access_key: ((artifacts.secret_access_key))
    regexp: ardumower-relay/chart/ardumower-relay-(.*).tgz

- name: image
  type: docker-manifest
  icon: floppy-variant
  check_every: 24h
  source:
    <<: *docker_image_source

- name: release
  type: github-release
  icon: folder-open
  source:
    owner: ((config.github.owner))
    repository: ((config.github.repository))
    access_token: ((github.access_token))

- name: dev
  <<: *helm
  source:
    token: ((dev.cluster_token))
    cluster_ca: ((dev.cluster_ca))
    cluster_url: ((dev.cluster_url))
    namespace: ((dev.namespace))
    release: ((dev.release))

- name: prod
  <<: *helm
  source:
    token: ((prod.cluster_token))
    cluster_ca: ((prod.cluster_ca))
    cluster_url: ((prod.cluster_url))
    namespace: ((prod.namespace))
    release: ((prod.release))

- name: go-build-image
  type: registry-image
  icon: hammer
  check_every: 24h
  source:
    repository: golang

- name: chart-build-image
  type: registry-image
  icon: hammer
  check_every: 24h
  source:
    repository: ghcr.io/typositoire/concourse-helm3-resource
    tag: v1.20.0

- name: semver-bumper
  <<: *github_artifact
  source:
    <<: *github_artifact_source
    owner: timotto
    repository: semver-bumper

resource_types:
- name: docker-manifest
  type: registry-image
  source:
    repository: mbialon/concourse-docker-manifest-resource
    tag: latest

- name: helm
  type: registry-image
  source:
    repository: ghcr.io/typositoire/concourse-helm3-resource
    tag: v1.20.0
  defaults:
    stable_repo: "false"

jobs:
- name: app
  serial_groups: [app]
  plan:
  - in_parallel:
    - get: app-source
      trigger: true
    - get: go-build-image
    - get: semver-bumper
      params: {globs: [semver-bumper-linux-amd64]}

  - in_parallel:
      fail_fast: true
      steps:
      - task: bump
        config:
          <<: *alpine_task
          inputs:
          - name: semver-bumper
          - name: app-source
          outputs:
          - name: bump
          run:
            path: sh
            args:
            - -exc
            - |
              mkdir bin
              export PATH=$PATH:$PWD/bin
              cp semver-bumper/semver-bumper-* bin/semver-bumper
              chmod +x bin/semver-bumper

              semver-bumper \
                -i cmd -i internal -i tools -i go.mod -i go.sum \
                -o "bump/release-version" \
                -c "bump/commitlog" \
                -t v \
                app-source

              semver-bumper \
                -i cmd -i internal -i tools -i go.mod -i go.sum \
                --pre rc \
                -o "bump/rc-version" \
                -t v \
                app-source

      - task: test
        image: go-build-image
        config:
          platform: linux
          inputs:
          - name: app-source
          caches:
          - path: cache/gopath
          run:
            path: sh
            args:
            - -exc
            - |
              export GOPATH=$PWD/cache/gopath

              cd app-source
              go test -race -cover ./...

  - load_var: app-version-val
    file: bump/rc-version
    format: trim
  - in_parallel:
      limit: 4
      fail_fast: true
      steps:
{{- range .platforms }}
      - do:
        - task: {{template "build" . }}
          output_mapping:
            build: {{template "build" . }}
          params:
            GOOS: {{.os}}
            GOARCH: {{.arch}}
            {{template "goArmParam" .}}
          <<: *build_task
{{- if .docker }}
        - task: {{template "image" . }}
          input_mapping:
            build: {{template "build" . }}
          output_mapping:
            image: {{template "image" . }}
          params:
            IMAGE_PLATFORM: {{template "imagePlatform" .}}
          <<: *image_task
        - put: {{template "image" . }}
          params:
            version: ((.:app-version-val))
            image: {{template "image" . }}/image.tar
          get_params:
            skip_download: true
{{- end}}
{{- end}}

  - put: image
    params:
      tag_file: bump/rc-version
      manifests:
{{- range .platforms }}
{{- if .docker}}
      - os: {{.os}}
        arch: {{.arch}}
        {{template "manifestPlatformVariant" .}}
        digest_file: {{template "image" . }}/digest
{{- end}}
{{- end}}

  - task: collect
    config:
      <<: *alpine_task
      inputs:
      - name: app-source
      - name: bump
        path: artifacts/bump
{{- range .platforms }}
      - name: {{template "build" . }}
{{- if .docker}}
      - name: {{template "image" . }}
{{- end}}
{{- end}}
      outputs:
      - name: result
      params:
        DOCKER_REPOSITORY: ((config.docker.repository))
      run:
        path: sh
        args:
        - -exc
        - |
          mkdir -p artifacts \
            artifacts/executables \
            artifacts/image_digests

          cat app-source/.git/short_ref | tee artifacts/git-ref
          release_version=$(cat artifacts/bump/release-version)
          echo "v$release_version" | tee artifacts/bump/release-name

          cat > artifacts/bump/release-body <<EOT
          Changes:
          EOT
          cat >> artifacts/bump/release-body < artifacts/bump/commitlog

          cp -v build-*/relay-* artifacts/executables/
          for image in image-*
          do
            cat ${image}/digest > artifacts/image_digests/${image}
          done

          (cd artifacts/executables ; sha256sum * | tee sha256sum.txt ; )

          cat >> artifacts/bump/release-body <<EOT

          Docker image:
          \`${DOCKER_REPOSITORY}:${release_version}\`

          Checksums:
          \`\`\`
          EOT
          cat >> artifacts/bump/release-body < artifacts/executables/sha256sum.txt
          cat >> artifacts/bump/release-body <<EOT
          \`\`\`
          EOT

          tar -cvzf result/result.tgz -C artifacts .
          du -sh result/result.tgz

  - put: app-artifacts
    params:
      file: result/result.tgz

  - put: app-source
    params:
      repository: app-source
      tag: bump/rc-version
      tag_prefix: v
      only_tag: true
      rebase: true

- name: chart
  serial_groups: [chart]
  plan:
  - in_parallel:
    - get: chart-source
      trigger: true
    - get: chart-build-image
    - get: semver-bumper
      params: {globs: [semver-bumper-linux-amd64]}

  - in_parallel:
    - task: bump
      config:
        <<: *alpine_task
        inputs:
        - name: semver-bumper
        - name: chart-source
        outputs:
        - name: bump
        run:
          path: sh
          args:
          - -exc
          - |
            mkdir bin
            export PATH=$PATH:$PWD/bin
            cp semver-bumper/semver-bumper-* bin/semver-bumper
            chmod +x bin/semver-bumper

            semver-bumper \
              -i helm \
              -o "bump/release-version" \
              -c "bump/commitlog" \
              -t v \
              chart-source

            semver-bumper \
              -i helm \
              --pre rc \
              -o "bump/rc-version" \
              -t v \
              chart-source

    - task: test
      image: chart-build-image
      config:
        platform: linux
        inputs:
        - name: chart-source
        run:
          path: sh
          args:
          - -ec
          - |
            src=chart-source/helm/ardumower-relay

            helm lint $src

  - task: bundle
    config:
      <<: *alpine_task
      inputs:
      - name: chart-source
      - name: bump
        path: artifacts/bump
      outputs:
      - name: result
      run:
        path: sh
        args:
        - -exc
        - |
          mkdir -p artifacts

          cat chart-source/.git/short_ref | tee artifacts/git-ref
          release_version=$(cat artifacts/bump/release-version)
          echo "v$release_version" | tee artifacts/bump/release-name

          cat > artifacts/bump/release-body <<EOT
          Changes:
          EOT
          cat >> artifacts/bump/release-body < artifacts/bump/commitlog

          tar -cvzf result/result.tgz -C artifacts .
          du -sh result/result.tgz

  - put: chart-artifacts
    params:
      file: result/result.tgz

  - put: chart-source
    params:
      repository: chart-source
      tag: bump/rc-version
      tag_prefix: chart-v
      only_tag: true
      rebase: true

- name: bundle
  serial_groups: [app,chart]
  plan:
  - in_parallel:
    - get: app-source
      passed: [app]
    - get: app-artifacts
      trigger: true
      passed: [app]
      params: {unpack: true}
    - get: chart-source
      passed: [chart]
    - get: chart-artifacts
      trigger: true
      passed: [chart]
      params: {unpack: true}
    - get: chart-build-image

  - task: chart
    image: chart-build-image
    config:
      platform: linux
      inputs:
      - name: chart-source
      - name: app-artifacts
      - name: chart-artifacts
      outputs:
      - name: chart
      run:
        path: sh
        args:
        - -ec
        - |
          app_version=$(cat app-artifacts/bump/rc-version)
          chart_version=$(cat chart-artifacts/bump/rc-version)
          final_app_version=$(cat app-artifacts/bump/release-version)
          final_chart_version=$(cat chart-artifacts/bump/release-version)
          src=chart-source/helm/ardumower-relay

          cp -vr chart-artifacts/bump chart/bump

          helm lint $src

          build() {
            bump="$1"
            app="$2"
            chart="$3"

            sed -e's/^appVersion:.*$/appVersion: "'"$app"'"/' -i $src/Chart.yaml
            sed -e's/^version:.*$/version: '"$chart"'/' -i $src/Chart.yaml

            mkdir -p build/$bump
            helm package --destination chart/$bump --version "$chart" $src
          }

          build rc "$app_version" "$chart_version"
          build release "$final_app_version" "$final_chart_version"
          mkdir chart/bundled
          cp -v chart/release/ardumower-relay-*tgz chart/bundled/ardumower-relay-chart.tgz

  - task: bundle
    config:
      <<: *alpine_task
      inputs:
      - name: app-artifacts
        path: src/app
      - name: chart
        path: src/chart
      outputs:
      - name: bundle
      run:
        path: tar
        args:
        - -czvf
        - bundle/bundle.tgz
        - -C
        - src
        - .

  - put: artifacts
    params:
      file: bundle/bundle.tgz

- name: test
  serial: true
  plan:
  - in_parallel:
    - get: app-source
      passed: [bundle]
      trigger: true
    - get: chart-source
      passed: [bundle]
      trigger: true
    - get: artifacts
      passed: [bundle]
      trigger: true
      params: {unpack: true}
    - get: smoketest
    - get: go-build-image

  - task: chart-values
    <<: *values_task
    params:
      VALUES_YAML: ((dev.values_yaml))
  - put: dev
    params:
      chart: artifacts/chart/rc/ardumower-relay-*.tgz
      timeout: 2m
      check_is_ready: true
      show_diff: true
      replace: true
      values: values/values.yaml

  - <<: *smoketest_task
    params:
      RELAY_SMOKETEST_SERVER_URL: ((dev.smoketest_url))
      RELAY_SMOKETEST_USERNAME: ((dev.smoketest_user))
      RELAY_SMOKETEST_PASSWORD: ((dev.smoketest_password))

- name: release
  serial: true
  plan:
  - in_parallel:
    - get: app-source
      passed: [test]
    - get: chart-source
      passed: [test]
    - get: artifacts
      passed: [test]

- name: release-app
  serial: true
  plan:
  - in_parallel:
    - get: app-source
      passed: [release]
      trigger: true
    - get: artifacts
      passed: [release]
      params: {unpack: true}

  - put: image
    params:
      tag_file: artifacts/app/bump/release-version
      manifests:
{{- range .platforms }}
  {{- if .docker}}
      - os: {{.os}}
        arch: {{.arch}}
        {{template "manifestPlatformVariant" .}}
        digest_file: artifacts/app/image_digests/{{template "image" .}}
  {{- end}}
{{- end}}
  - put: app-source
    params:
      repository: app-source
      tag: artifacts/app/bump/release-version
      tag_prefix: v
      only_tag: true
  - put: release
    params:
      name: artifacts/app/bump/release-name
      body: artifacts/app/bump/release-body
      tag: artifacts/app/bump/release-version
      tag_prefix: v
      globs:
      - artifacts/app/executables/relay-*
      - artifacts/app/executables/sha256sum.txt
      - artifacts/chart/bundled/ardumower-relay-chart.tgz
    get_params:
      globs: [none]

- name: release-chart
  serial: true
  plan:
  - in_parallel:
    - get: chart-source
      trigger: true
      passed: [release]
    - get: artifacts
      passed: [release]
      params: {unpack: true}

  - put: chart-artifact
    params:
      file: artifacts/chart/release/ardumower-relay-*.tgz

  - put: chart-source
    params:
      repository: chart-source
      tag: artifacts/chart/bump/release-version
      tag_prefix: chart-v
      only_tag: true

- name: production
  serial: true
  plan:
  - in_parallel:
    - get: release
      trigger: true
      passed: [release-app]
      params: {globs: [ardumower-relay-*]}
    - get: smoketest
    - get: go-build-image
  - task: chart-values
    <<: *values_task
    params:
      VALUES_YAML: ((prod.values_yaml))
  - put: prod
    params:
      chart: release/ardumower-relay-*.tgz
      timeout: 3m
      check_is_ready: true
      show_diff: true
      replace: true
      values: values/values.yaml
  - <<: *smoketest_task
    params:
      RELAY_SMOKETEST_SERVER_URL: ((prod.smoketest_url))
      RELAY_SMOKETEST_USERNAME: ((prod.smoketest_username))
      RELAY_SMOKETEST_PASSWORD: ((prod.smoketest_password))

- name: pipeline
  serial: true
  plan:
  - in_parallel:
    - get: ci-source
      trigger: true
    - get: go-build-image
  - task: generate
    image: go-build-image
    config:
      platform: linux
      inputs:
      - name: ci-source
        path: .
      outputs:
      - name: pipeline
        path: .
      run:
        path: go
        args:
        - generate
        - ./ci
  - set_pipeline: self
    file: pipeline/ci/pipeline.yml
    vars:
      config:
        github:
          owner: ((config.github.owner))
          repository: ((config.github.repository))
          uri: ((config.github.uri))
          branch: ((config.github.branch))
        docker:
          repository: ((config.docker.repository))

groups:
- name: everything
  jobs:
  - app
  - chart
  - bundle
  - test
  - release
  - release-app
  - release-chart
  - production
  - pipeline
- name: dev
  jobs:
  - app
  - chart
  - bundle
  - test
  - pipeline
- name: release
  jobs:
  - release
  - release-app
  - release-chart
  - production
