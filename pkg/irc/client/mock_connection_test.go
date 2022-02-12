package client

import "io"

type MockConn struct {
	ServerReader *io.PipeReader
	ServerWriter *io.PipeWriter

	ClientReader *io.PipeReader
	ClientWriter *io.PipeWriter
}

func (c MockConn) Close() error {
	if err := c.ServerWriter.Close(); err != nil {
		return err
	}
	return c.ServerReader.Close()
}

func (c MockConn) Read(data []byte) (int, error) {
	return c.ServerReader.Read(data)
}

func (c MockConn) Write(data []byte) (int, error) {
	return c.ServerWriter.Write(data)
}

func MakeMockConn() MockConn {
	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()
	return MockConn{
		ServerReader: serverRead,
		ServerWriter: serverWrite,
		ClientReader: clientRead,
		ClientWriter: clientWrite,
	}
}
