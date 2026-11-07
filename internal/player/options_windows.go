package player

// platformOptions are base options for this platform. The window draws
// with OpenGL through ANGLE, which shows hardware-decoded frames through
// d3d11-egl; mpv otherwise loads every interop it has (Vulkan, CUDA,
// ...) on the first video, for memory and nothing else.
var platformOptions = []Option{
	{"gpu-hwdec-interop", "d3d11-egl"},
}
