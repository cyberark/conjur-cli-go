#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
    timeout(time: 2, unit: 'HOURS')
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  stages {
    stage('Run Unit Tests') {
      steps {
        sh './bin/test_unit'

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
          ccCoverage("gocov", "--prefix github.com/cyberark/conjur-cli-go")
      }
    }
    stage('Run Integration Tests') {
      steps {
        dir('ci') {
          script {
            try{
              sh 'summon -f ./okta/secrets.yml ./test_integration'
            } finally {
              archiveArtifacts 'cleanup.log'
            }
          }
        }
      }
    }
  }
}
