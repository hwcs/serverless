package model

type Code struct {
    ZipFile []byte
}

type Function struct {
    FuncCode  Code
    Description string
}

type EventMapping struct {
    BatchSize int
}


     
   
