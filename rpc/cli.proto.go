syntax = "proto3";

package rpc


service Cli {
rpc PublishTx (Transaction) returns (Response) {}
rpc GetTransaction (TransactionKey) returns (Transaction) {}
rpc GetBalance (Key) returns (Value){}
rpc GetState (Key) returns (Value){}
rpc GetBlock (BlockKey) returns (BlockInfo){}
}

message Transaction {
bytes tx = 1;
}

message Response {
int32 code = 1;
}

message TransactionKey {
string publisher = 1;
int64 nonce = 2;
}

message Key {
string s = 1;
}

message Value {
string sv = 2;
}

message BlockKey {
int64 layer = 1;
}

message BlockInfo {
string json = 1;
}