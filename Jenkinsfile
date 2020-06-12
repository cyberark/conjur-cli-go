#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  stages {
    stage('Check Changelog') {
      steps {
        sh './bin/check_changelog'
      }
    }

    stage('Run tests') {
      steps {
        sh './bin/test'
      }
      post {
        always {
          junit 'output/junit.xml'
        }
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
