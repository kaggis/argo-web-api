pipeline {
    agent { 
        docker { 
            image 'epel7mgo' 
            label 'slave01'
            args '-u jenkins:jenkins'
        }
    }
    options { checkoutToSubdirectory('argo-web-api') }
    environment {
        PROJECT_DIR='argo-web-api'
        GOPATH="${WORKSPACE}/go"
        GIT_COMMIT=sh(script: 'cd ${WORKSPACE}/argo-web-api && git log -1 --format="%H"',returnStdout: true).trim()
        GIT_COMMIT_HASH=sh(script: 'cd ${WORKSPACE}/argo-web-api && git log -1 --format="%H" | cut -c1-7',returnStdout: true).trim()
        GIT_COMMIT_DATE=sh(script: 'date -d "$(cd ${WORKSPACE}/argo-web-api && git show -s --format=%ci ${GIT_COMMIT_HASH})" "+%Y%m%d%H%M%S"',returnStdout: true).trim()
    }
    stages {
         
        stage('Build') {
            
            steps {
                echo 'Build...'
                sh """
                mkdir -p ${WORKSPACE}/go/src/github.com/ARGOeu
                ln -sf ${WORKSPACE}/${PROJECT_DIR} ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                rm -rf ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}/${PROJECT_DIR}
                cd ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                go build
                """
            }
        }
        stage('Test') {
            steps {
                echo 'Test & Coverage...'
                sh """
                sudo /etc/init.d/mongod restart
                cd ${WORKSPACE}/go/src/github.com/ARGOeu/${PROJECT_DIR}
                gocov test -p 1 \$(go list ./... | grep -v /vendor/) | gocov-xml > ${WORKSPACE}/coverage.xml
                go test -p 1 \$(go list ./... | grep -v /vendor/) -v=1 | go-junit-report > ${WORKSPACE}/junit.xml
                """
                junit '**/junit.xml'
                cobertura coberturaReportFile: '**/coverage.xml'

            }
        }
        stage('Package') {
            steps {
                echo 'Building Rpm...'
                sh """
                cd ${WORKSPACE}/${PROJECT_DIR} && make sources
                cp ${WORKSPACE}/${PROJECT_DIR}/argo-web-api*.tar.gz /home/jenkins/rpmbuild/SOURCES/
                if [ "$env.BRANCH_NAME" != "master" ]; then
                    sed -i 's/^Release.*/Release: %(echo $GIT_COMMIT_DATE).%(echo $GIT_COMMIT_HASH)%{?dist}/' ${WORKSPACE}/${PROJECT_DIR}/argo-web-api.spec
                fi
                cd /home/jenkins/rpmbuild/SOURCES && tar -xzvf argo-web-api*.tar.gz
                cp ${WORKSPACE}/${PROJECT_DIR}/argo-web-api.spec /home/jenkins/rpmbuild/SPECS/
                rpmbuild -bb /home/jenkins/rpmbuild/SPECS/*.spec
                rm -f ${WORKSPACE}/*.rpm
                cp /home/jenkins/rpmbuild/RPMS/**/*.rpm ${WORKSPACE}/
                """
                archiveArtifacts artifacts: '**/*.rpm', fingerprint: true
                script {
                    if ( env.BRANCH_NAME == 'master' ) {
                        echo 'Uploading rpm for devel...'
                        withCredentials(bindings: [sshUserPrivateKey(credentialsId: 'jenkins-repo', usernameVariable: 'REPOUSER', \
                                                                keyFileVariable: 'REPOKEY')]) {
                            sh  '''
                                scp -i ${REPOKEY} -o StrictHostKeyChecking=no ${WORKSPACE}/*.rpm ${REPOUSER}@rpm-repo.argo.grnet.gr:/repos/ARGO/prod/centos7/
                                ssh  -i ${REPOKEY} -o StrictHostKeyChecking=no ${REPOUSER}@rpm-repo.argo.grnet.gr createrepo --update /repos/ARGO/prod/centos7/
                                '''
                        }
                    }
                    else if ( env.BRANCH_NAME == 'devel' ) {
                        echo 'Uploading rpm for devel...'
                        withCredentials(bindings: [sshUserPrivateKey(credentialsId: 'jenkins-repo', usernameVariable: 'REPOUSER', \
                                                                    keyFileVariable: 'REPOKEY')]) {
                            sh  '''
                                scp -i ${REPOKEY} -o StrictHostKeyChecking=no ${WORKSPACE}/*.rpm ${REPOUSER}@rpm-repo.argo.grnet.gr:/repos/ARGO/devel/centos7/
                                ssh -i ${REPOKEY} -o StrictHostKeyChecking=no ${REPOUSER}@rpm-repo.argo.grnet.gr createrepo --update /repos/ARGO/devel/centos7/
                                '''
                        }
                    }
                }
            }
        }
        stage('Package for Production') {
            when {
                branch 'master'
            }
            steps {
                echo 'Building Rpm for Production...'
                sh '''cd ${WORKSPACE}/argo-web-api && make sources
                cd /home/jenkins/rpmbuild/SOURCES && tar -xzvf argo-web-api*.tar.gz
                cp ${WORKSPACE}/argo-web-api/argo-web-api.spec /home/jenkins/rpmbuild/SPECS/
                rpmbuild -bb /home/jenkins/rpmbuild/SPECS/*.spec
                rm -f ${WORKSPACE}/*.rpm
                cp /home/jenkins/rpmbuild/RPMS/**/*.rpm ${WORKSPACE}/
                '''
                archiveArtifacts artifacts: '**/*.rpm', fingerprint: true
                echo 'Uploading rpm for devel...'
                withCredentials(bindings: [sshUserPrivateKey(credentialsId: 'jenkins-repo', usernameVariable: 'REPOUSER', \
                                                             keyFileVariable: 'REPOKEY')]) {
                  sh '''
                  scp -i ${REPOKEY} -o StrictHostKeyChecking=no ${WORKSPACE}/*.rpm ${REPOUSER}@rpm-repo.argo.grnet.gr:/repos/ARGO/prod/centos7/
                  '''
                }
            }
        }
        
    }
}