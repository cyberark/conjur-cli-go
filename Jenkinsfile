#!/usr/bin/env groovy
@Library([
  "product-pipelines-shared-library",
  "conjur-enterprise-sharedlib"
]) _

def productName = 'Conjur CLI'
def productTypeName = 'Conjur Enterprise'

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([
    'conjur-enterprise/conjur-api-go',
  ])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { INFRAPOOL_EXECUTORV2_AGENT_0, sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    env.INFRAPOOL_PRODUCT_NAME = "${productName}"
    env.INFRAPOOL_DD_PRODUCT_TYPE_NAME = "${productTypeName}"

    // Scan the image before promoting
    // NOTE: no-handle form per migration-patterns.md#8 — promote() allocates its own
    // ExecutorV2 with no local Snyk output to sync, so the Snyk->DefectDojo sync
    // silently no-ops here regardless of handle. Signature update only, not an
    // execution rewrite of the promote-closure agentSh/agentPut/agentGet calls below.
    runSecurityScans(
      image: "registry.tld/conjur-cli:${sourceVersion}-${gitCommit(INFRAPOOL_EXECUTORV2_AGENT_0)}",
      buildMode: params.MODE,
      branch: env.BRANCH_NAME,
      arch: 'linux/amd64'
    )

    //Sign *.deb, *.exe, *.tar.gz and conjur_darwin_* artifacts
    INFRAPOOL_EXECUTORV2_AGENT_0.agentGet from: "${assetDirectory}/", to: "./"

    signArtifacts patterns: ["*.tar.gz"]
    signArtifacts patterns: ["conjur_darwin_*"]
    signArtifacts patterns: ["*.deb"]
    signArtifacts patterns: ["*.exe"]

    INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "*.sig", to: "${assetDirectory}"
    INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "*.deb", to: "${assetDirectory}"
    INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "*.exe", to: "${assetDirectory}"
    INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "conjur_darwin_*", to: "${assetDirectory}"

    // Promote source version to target version.

    // NOTE: the use of --pull to ensure source images are pulled from internal registry
    INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "source ./bin/build_utils && ./bin/publish_container_images --promote --source ${sourceVersion}-\$(git_commit) --target ${targetVersion} --pull"

    // Ensure the working directory is a safe git directory for the subsequent
    // promotion operations after this block.
    sh 'git config --global --add safe.directory "$(pwd)"'
  }

  // Copy Github Enterprise release to Github
  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-AmznDocker' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
    timeout(time: 1, unit: 'HOURS')
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
    // Ensures CI uses the internal registry for conjur edge images
    REGISTRY_URL = "registry.tld"
  }

  triggers {
    cron(getDailyCronString())
    parameterizedCron(getWeeklyCronString("H(1-5)","%MODE=RELEASE"))
  }

  stages {
    // Aborts any builds triggered by another project that wouldn't include any changes
    stage ("Skip build if triggering job didn't create a release") {
      when {
        expression {
          MODE == "SKIP"
        }
      }
      steps {
        script {
          currentBuild.result = 'ABORTED'
          error("Aborting build because this build was triggered from upstream, but no release was built")
        }
      }
    }

    stage('Scan for internal URLs') {
      steps {
        script {
          detectInternalUrls()
        }
      }
    }

    stage('Mark Workspace as Safe Git Directory') {
      steps {
        script {
          sh "git config --global --add safe.directory $WORKSPACE"
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate changelog and set version') {
      steps {
        updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        script {
          updatePrivateGoDependencies("${WORKSPACE}/go.mod")
        }
      }
    }

    stage('Build while unit testing') {
      parallel {
        stage('Run unit tests') {
          steps {
            script {
              sh './bin/test_unit'
            }
          }
          post {
            always {
              script {
                sh './bin/coverage'
                junit 'junit.xml'

                recordCoverage(
                  tools: [
                    [parser: 'COBERTURA', pattern: 'coverage.xml']
                  ],
                  sourceCodeEncoding: 'UTF-8',
                  enabledForFailure: true,
                  qualityGates: [
                    [
                      metric: 'LINE',
                      threshold: 70.0,
                      criticality: 'UNSTABLE'
                    ],
                    [
                      metric: 'METHOD',
                      threshold: 70.0,
                      criticality: 'UNSTABLE'
                    ],
                    [
                      metric: 'BRANCH',
                      threshold: 0.0,
                      criticality: 'UNSTABLE'
                    ]
                  ]
                )
                catchError(buildResult: 'SUCCESS', stageResult: 'UNSTABLE') {
                  codacy action: 'reportCoverage', filePath: "coverage.xml"
                }
              }
            }
          }
        }

        stage('Build release artifacts') {
          steps {
            script {
              dir('./pristine-checkout') {
                // Go releaser requires a pristine checkout
                checkout scm

                // Copy VERSION info into prisitine folder
                sh "cp ../VERSION ./VERSION"

                // Create release artifacts without releasing to Github
                sh "./bin/build_release --skip=validate --clean"

                // Build container images
                sh "./bin/build_container_images"

                // Archive release artifacts
                archiveArtifacts(
                  artifacts: [
                    'dist/conjur',
                    'dist/goreleaser/binaries/**',
                    'dist/goreleaser/**/*.tar.gz',
                    'dist/goreleaser/**/*.zip',
                    'dist/goreleaser/**/*.deb',
                    'dist/goreleaser/**/*.rpm',
                    'dist/goreleaser/**/*.exe',
                    'dist/goreleaser/**/*.sha256',
                    'dist/goreleaser/**/*.txt',
                    'dist/goreleaser/**/*.json',
                    'dist/goreleaser/**/*.yml',
                    'dist/goreleaser/**/*.yaml'
                  ].join(','),
                  allowEmptyArchive: true
                )
              }
            }
          }
        }

        stage('Run integration tests for huh in TTY') {
          steps {
            script {
              try {
                sh './ci/test_vhs'
              } finally {
                archiveArtifacts artifacts: 'vhs/output/*'
              }
            }
          }
        }
      }
    }

    // Publish container images to internal registry. Need to push before we do security scans
    // since the Snyk scans pull from artifactory on a seprate executor node
    stage('Push images to internal registry') {
      steps {
        script {
          sh './bin/publish_container_images --internal'
        }
      }
    }

    stage('Integration test while scanning') {
      environment {
        INFRAPOOL_REGISTRY_URL = "${REGISTRY_URL}"
        // Values to direct scan results to the right place in DefectDojo
        INFRAPOOL_PRODUCT_NAME = "${productName}"
        INFRAPOOL_DD_PRODUCT_TYPE_NAME = "${productTypeName}"
      }
      parallel {
        stage('Run integration tests') {
          steps {
            withCredentials([
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_Operating_System-WindowsDomainAccountDailyRotation-cyberng.com-svc_cnjr_enterprise_username", variable: 'INFRAPOOL_IDENTITY_USERNAME'),
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_Operating_System-WindowsDomainAccountDailyRotation-cyberng.com-svc_cnjr_enterprise_password", variable: 'INFRAPOOL_IDENTITY_PASSWORD')
            ])
            {
              script {
                dir('ci') {
                  try {
                    sh 'summon -f ./secrets.yml -e ci ./test_integration'
                  } finally {
                    // Relative to dir('ci') above, unlike the removed infrapool
                    // handle's agentArchiveArtifacts, which resolved from the
                    // orchestrator workspace root (hence the old 'ci/' prefix).
                    archiveArtifacts artifacts: 'cleanup.log'
                  }
                }
              }
            }
          }
        }

        stage("Scan main Docker image") {
          steps {
            script {
              runSecurityScans(
                image: "registry.tld/${containerImageWithTag()}",
                buildMode: params.MODE,
                branch: env.BRANCH_NAME,
                arch: 'linux/amd64'
              )
            }
          }
        }

        stage('Scan with govulncheck') {
          steps {
            script {
              sh 'go install golang.org/x/vuln/cmd/govulncheck@latest'
              catchError(buildResult: 'UNSTABLE', stageResult: 'UNSTABLE') {
                sh '$(go env GOPATH)/bin/govulncheck -mode=binary ./pristine-checkout/dist/goreleaser/binaries/conjur_linux_amd64_v1'
              }
            }
          }
        }
      }
    }

    // TODO: Uncomment this after 9.2.3 release
    // stage('Run Secrets Manager SaaS tests') {
    //   stages {
    //     stage('Create a Tenant') {
    //       steps {
    //         script {
    //           TENANT = getConjurCloudTenant()
    //         }
    //       }
    //     }
    //     stage('Run tests against Tenant') {
    //       environment {
    //         INFRAPOOL_CONJUR_APPLIANCE_URL="${TENANT.conjur_cloud_url}"
    //         INFRAPOOL_IDENTITY_USERNAME_CLOUD="${TENANT.login_name}"
    //       }
    //       steps {
    //         script {
    //           INFRAPOOL_EXECUTORV2_AGENT_0.agentDir('ci') {
    //             try {
    //               INFRAPOOL_EXECUTORV2_AGENT_0.agentSh 'summon -f ./secrets.yml -e ci ./test_integration_cloud'
    //             } finally {
    //               INFRAPOOL_EXECUTORV2_AGENT_0.agentArchiveArtifacts artifacts: 'cloud_cleanup.log'
    //             }
    //           }
    //         }
    //       }
    //     }
    //   }
    //   post {
    //     always {
    //       script {
    //         deleteConjurCloudTenant("${TENANT.id}")
    //       }
    //     }
    //   }
    // }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        script {
          release { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
            // Publish release artifacts to all the appropriate locations
            // Copy any artifacts to assetDirectory to attach them to the Github release

            // Copy assets to be published in Github release.
            sh "${toolsDirectory}/bin/copy_goreleaser_artifacts ${assetDirectory}"

            // Create Go application SBOM using the go.mod version for the golang container image
            sh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/conjur/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
            // Create Go module SBOM
            sh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """
          }
        }
      }
    }
  }
}

def gitCommit(infrapool) {
  infrapool.agentSh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "$(git_commit)"'
  )
}

def containerImageWithTag() {
  sh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "conjur-cli:$(project_version_with_commit)"'
  )
}
