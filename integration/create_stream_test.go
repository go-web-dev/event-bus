//+build integration

package integration

func (s *appSuite) Test_CreateStream_Success() {
	conn := s.newConn()
	s.write(conn, "health", "")
}
