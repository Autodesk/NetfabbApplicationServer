@Library('PSL')
import ors.utils.common_scm
import ors.ci.common_ci


winNode = "netfabb_build_win10_2"

String cron_string = BRANCH_NAME == "master" ? "11 11 * * *" : ""
properties (
    [
        pipelineTriggers([[$class: 'TimerTrigger', spec: cron_string]])
    ]
)

@NonCPS
def artifactoryUpload(String fileToUpload)
{    
    common_artifactory=new ors.utils.common_artifactory(steps,env,Artifactory,'svc_p_netfabbjenkins')
    def uploadSpec = """{
        "files": [
                {
                    "pattern": "${fileToUpload}",
                    "target": "team-netfabb-generic/NetfabbApplicationServer/${env.BRANCH_NAME}/"
                }
            ]
        }"""
    try {
      common_artifactory.upload('https://art-bobcat.autodesk.com/artifactory/',uploadSpec)
    } catch (Exception e){
      echo "ERROR: Caught exception in artifactoryUpload: " + e
    }        
}

def sign(String fileToSign){
  withCredentials([string(credentialsId: 'OTP_KEY', variable: 'OTP_KEY')]) {
    bat "C:\\signing\\sign.bat ${fileToSign} ${OTP_KEY}"
  }
}


pipeline {
    agent none
    options {
        disableConcurrentBuilds()
        skipDefaultCheckout true
    }
    stages{
        stage ('Build') {
            parallel {
                stage ('Build Windows') {
                    agent {
                        node {
                            label winNode
                            customWorkspace 'E:\\NAS'
                        }
                    }
                    stages {
                        stage ('Update GIT') {
                            steps {
                                checkout scm
                                bat 'git fetch upstream'
                                bat 'git merge upstream/master'
                            }
                        }
                        stage ('Build') {
                            steps {
                                bat 'build_jenkins.bat'
                            }
                        }
                        stage ('Signing') {
                            steps {
                                sign(pwd() + '\\output\\NetfabbApplicationServer.exe')
                                sign(pwd() + '\\output\\NetfabbApplicationService.exe')
                                sign(pwd() + '\\output\\PasswordSalter.exe')
                            }
                        }
                        stage ('Artifactory') {
                            steps {
                                script {
                                    artifactoryUpload('/output/Documentation/Netfabb_Application_Server.docx')
                                    artifactoryUpload('/output/Documentation/Netfabb_Application_Server.pdf')
                                    artifactoryUpload('/output/example.crt')
                                    artifactoryUpload('/output/example.key')
                                    artifactoryUpload('/output/favicon.ico')
                                    artifactoryUpload('/output/netfabbapplicationserver.db')
                                    artifactoryUpload('/output/netfabbapplicationserver.db.shipping')
                                    artifactoryUpload('/output/NetfabbApplicationServer.exe')
                                    artifactoryUpload('/output/netfabbapplicationserver.xml')
                                    artifactoryUpload('/output/NetfabbApplicationService.exe')
                                    artifactoryUpload('/output/netfabbormschemas.json')
                                    artifactoryUpload('/output/netfabbtasks.db')
                                    artifactoryUpload('/output/PasswordSalter.exe')
                                    artifactoryUpload('/output/setup_firewall_rules.bat')
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
