#!/usr/bin/env bash



# -p 并发数
# -n 执行次数
# -v 打印输出
python dstest.py -v  -p 20 -n 100 -o ./output TestPersist12C TestPersist22C TestPersist32C TestFigure82C TestUnreliableAgree2C TestFigure8Unreliable2C TestReliableChurn2C TestUnreliableChurn2C

python dstest.py -v  -p 15 -n 50 -o ./output TestSnapshotBasic2D

python dstest.py -v  -p 15 -n 50 -o ./output TestSnapshotInstall2D TestSnapshotInstallUnreliable2D TestSnapshotInstallCrash2D TestSnapshotInstallUnCrash2D


python ../common/dstest.py -v  -p 10 -n 100 -o ./output InitialElection2A TestReElection2A TestManyElections2A BasicAgree2B RPCBytes2B FailAgree2B FailNoAgree2B ConcurrentStarts2B Rejoin2B Backup2B Count2B TestPersist12C TestPersist22C TestPersist32C TestFigure82C TestUnreliableAgree2C TestFigure8Unreliable2C TestReliableChurn2C TestUnreliableChurn2C TestSnapshotBasic2D TestSnapshotInstall2D TestSnapshotInstallUnreliable2D TestSnapshotInstallCrash2D TestSnapshotInstallUnCrash2D

python dslog.py output/TestSnapshotBasic2D_13.log -c 3

InitialElection2A
TestReElection2A
TestManyElections2A
BasicAgree2B
RPCBytes2B
FailAgree2B
FailNoAgree2B
ConcurrentStarts2B
Rejoin2B
Backup2B
Count2B
TestPersist12C
TestPersist22C
TestPersist32C
TestFigure82C
TestUnreliableAgree2C
TestFigure8Unreliable2C
TestReliableChurn2C
TestUnreliableChurn2C
TestSnapshotBasic2D
TestSnapshotInstall2D
TestSnapshotInstallUnreliable2D
TestSnapshotInstallCrash2D
TestSnapshotInstallUnCrash2D