#!/usr/bin/env groovy
@Library("product-pipelines-shared-library") _

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
  release.promote(params.VERSION_TO_PROMOTE) { infrapool, sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    // Promote source version to target version.

    // NOTE: the use of --pull to ensure source images are pulled from internal registry
    infrapool.agentSh "source ./bin/build_utils && ./bin/publish_container_images --promote --source ${sourceVersion}-\$(git_commit) --target ${targetVersion} --pull"
  }

  // Copy Github Enterprise release to Github
  release.copyEnterpriseRelease(params.VERSION_TO_PROMOTE)
  return
}

pipeline {
  agent { label 'conjur-enterprise-common-agent' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
    timeout(time: 30, unit: 'MINUTES')
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  triggers {
    cron(getDailyCronString())
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

    stage('Get InfraPool ExecutorV2 Agent') {
      steps {
        script {
          // Request ExecutorV2 agents for 1 hour(s)
          INFRAPOOL_EXECUTORV2_AGENT_0 = getInfraPoolAgent.connected(type: "ExecutorV2", quantity: 1, duration: 1)[0]
        }
      }
    }

    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps { 
            script {
              INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/parse-changelog'
            }
          }
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate changelog and set version') {
      steps {
        updateVersion(INFRAPOOL_EXECUTORV2_AGENT_0, "CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        script {
          updatePrivateGoDependencies("${WORKSPACE}/go.mod")
          // Copy the vendor directory onto infrapool
          INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "vendor", to: "${WORKSPACE}"
          INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "go.*", to: "${WORKSPACE}"
          // Add GOMODCACHE directory to infrapool allowing automated release to generate SBOMs
          INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "/root/go", to: "/var/lib/jenkins/"
        }
      }
    }

    stage('Build while unit testing') {
      parallel {
        stage('Run unit tests') {
          steps {
            script {
              INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/test_unit'
            }
          }
          post {
            always {
              script {
                INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/coverage'
                INFRAPOOL_EXECUTORV2_AGENT_0.agentStash name: 'xml-out', includes: '*.xml'
                unstash 'xml-out'
                junit 'junit.xml'

                cobertura autoUpdateHealth: false,
                  autoUpdateStability: false,
                  coberturaReportFile: 'coverage.xml',
                  conditionalCoverageTargets: '70, 0, 0',
                  failUnhealthy: false,
                  failUnstable: false,
                  maxNumberOfBuilds: 0,
                  lineCoverageTargets: '70, 0, 0',
                  methodCoverageTargets: '70, 0, 0',
                  onlyStable: false,
                  sourceEncoding: 'ASCII',
                  zoomCoverageChart: false
                  codacy action: 'reportCoverage', filePath: "coverage.xml"
              }
            }
          }
        }

        stage('Build release artifacts') {
          steps {
            script {
              INFRAPOOL_EXECUTORV2_AGENT_0.agentDir('./pristine-checkout') {
                // Go releaser requires a pristine checkout
                checkout scm

                // Copy the checkout content onto infrapool
                INFRAPOOL_EXECUTORV2_AGENT_0.agentPut from: "./", to: "."

                // Copy VERSION info into prisitine folder
                INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "cp ../VERSION ./VERSION"

                // Create release artifacts without releasing to Github
                INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "./bin/build_release --skip-validate --clean"

                // Build container images
                INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "./bin/build_container_images"

                // Archive release artifacts
                INFRAPOOL_EXECUTORV2_AGENT_0.agentArchiveArtifacts artifacts: 'dist/goreleaser/'
              }
            }
          }
        }
      }
    }

    stage('Integration test while scanning') {
      parallel {
        stage('Run integration tests') {
          steps {
            withCredentials([
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_Operating_System-WindowsDomainAccountDailyRotation-cyberng.com-svc_cnjr_enterprise_username", variable: 'INFRAPOOL_IDENTITY_USERNAME'),
              conjurSecretCredential(credentialsId: "RnD-Global-Conjur-Ent-Conjur_Operating_System-WindowsDomainAccountDailyRotation-cyberng.com-svc_cnjr_enterprise_password", variable: 'INFRAPOOL_IDENTITY_PASSWORD')
            ]) 
            {
              script {
                INFRAPOOL_EXECUTORV2_AGENT_0.agentDir('ci') {
                  try {
                    INFRAPOOL_EXECUTORV2_AGENT_0.agentSh 'summon -f ./secrets.yml -e ci ./test_integration'
                  } finally {
                    INFRAPOOL_EXECUTORV2_AGENT_0.agentArchiveArtifacts artifacts: 'cleanup.log'
                  }
                }
              }
            }
          }
        }

        stage("Scan container images for fixable issues") {
          steps {
            script {
              scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, "${containerImageWithTag()}", "HIGH", false)
            }
          }
        }

        stage("Scan container images for total issues") {
          steps {
            script {
              scanAndReport(INFRAPOOL_EXECUTORV2_AGENT_0, "${containerImageWithTag()}", "NONE", true)
            }
          }
        }
      }
    }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        script {
          release(INFRAPOOL_EXECUTORV2_AGENT_0) { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
            // Publish release artifacts to all the appropriate locations
            // Copy any artifacts to assetDirectory to attach them to the Github release

            // Copy assets to be published in Github release.
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh "./bin/copy_release_artifacts ${assetDirectory}"

            // Create Go application SBOM using the go.mod version for the golang container image
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/conjur/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
            // Create Go module SBOM
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh """export PATH="${toolsDirectory}/bin:${PATH}" && go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """

            // Publish container images to internal registry
            INFRAPOOL_EXECUTORV2_AGENT_0.agentSh './bin/publish_container_images --internal'
          }
        }
      }
    }
  }

  post {
    always {
      script {
        releaseInfraPoolAgent(".infrapool/release_agents")
      }
    }
  }
}

def containerImageWithTag() {
  INFRAPOOL_EXECUTORV2_AGENT_0.agentSh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "conjur-cli:$(project_version_with_commit)"'
  )
}
