//+build integration

package integration

func (s *appSuite) Test_DeleteStream_Success() {
	conn := s.newConn()

	s.write(conn, "delete_stream", `{"stream_name": "existing-stream-name"}`)

	var res response
	s.read(conn, &res)
	s.Equal("delete_stream", res.Operation)
	s.True(res.Status)
	s.Empty(res.Body)
}
