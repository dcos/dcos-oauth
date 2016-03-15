// +build ignore

package integration

func testConfig(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(startZk())
	assert.NoError(startConfigAPI())
	defer cleanup("dcos-config")

	config := api.Config{
		ClusterConfiguration: struct {
			FirstUser bool
		}{
			FirstUser: true,
		},
	}
	body, err := send("GET", "/config", 404, nil)
	assert.NoError(err)
	assert.Equal("Config not found", body, "Config should be empty")

	_, err = send("PUT", "/config", 200, config)
	assert.NoError(err)

	body, err = send("GET", "/config", 200, nil)
	assert.NoError(err)
	assert.True(reflect.DeepEqual(body, config), "Config should match default config")

	_, err = send("PATCH", "/config", 200, nil)
	assert.NoError(err)

	body, err = send("GET", "/config", 200, nil)
	assert.NoError(err)
	assert.True(reflect.DeepEqual(body, config), "Config should match patched config")
}
