package com.cnp

import android.Manifest
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.opengl.Visibility
import android.os.*
import android.os.Handler
import android.util.Log
import android.view.View.GONE
import android.view.View.VISIBLE
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.camera.core.*
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.core.view.isVisible
import com.github.kittinunf.fuel.Fuel
import com.github.kittinunf.fuel.core.*
import com.github.kittinunf.fuel.httpGet
import com.github.kittinunf.fuel.httpUpload
import kotlinx.android.synthetic.main.activity_main.*
import kotlinx.coroutines.GlobalScope
import kotlinx.coroutines.async
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import java.io.*
import java.nio.MappedByteBuffer
import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors


class MainActivity : AppCompatActivity() {
    private var preview: Preview? = null
    private var imageCapture: ImageCapture? = null
    private var camera: Camera? = null
    private var serverAddress = "http://192.168.1.28"

    private lateinit var cameraExecutor: ExecutorService

    private var currentState = "image"

    private var imageWidth = 600
    private var imageHeight = 600

    public var h = Handler()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        setContentView(R.layout.activity_main)

        // Request camera permissions
        if (allPermissionsGranted()) {
            startCamera()
        } else {
            ActivityCompat.requestPermissions(
                this, REQUIRED_PERMISSIONS, REQUEST_CODE_PERMISSIONS)
        }

        // Setup the listener for take photo button
        viewFinder.setOnClickListener { takePhoto() }

        cameraExecutor = Executors.newSingleThreadExecutor()

    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == REQUEST_CODE_PERMISSIONS) {
            if (allPermissionsGranted()) {
                startCamera()
            } else {
                Toast.makeText(this,
                    "Permissions not granted by the user.",
                    Toast.LENGTH_SHORT).show()
                finish()
            }
        }
    }

    private fun startCamera() {
        val cameraProviderFuture = ProcessCameraProvider.getInstance(this)

        cameraProviderFuture.addListener(Runnable {
            val cameraProvider: ProcessCameraProvider = cameraProviderFuture.get()

            preview = Preview.Builder().build()

            imageCapture = ImageCapture.Builder().apply {
                setCaptureMode(ImageCapture.CAPTURE_MODE_MINIMIZE_LATENCY)
            }.build()

            val cameraSelector = CameraSelector.Builder().requireLensFacing(CameraSelector.LENS_FACING_BACK).build()

            try {
                cameraProvider.unbindAll()

                camera = cameraProvider.bindToLifecycle(
                    this, cameraSelector, preview, imageCapture)
                preview?.setSurfaceProvider(viewFinder.createSurfaceProvider())
            } catch (exc: Exception) {
                Log.e(TAG, "Use camera binding failed", exc)
            }

        }, ContextCompat.getMainExecutor(this))

    }

    fun takePhoto() {
        val imageCapture = imageCapture?: return

        imageCapture.takePicture(
            ContextCompat.getMainExecutor(this), object : ImageCapture.OnImageCapturedCallback() {
                override fun onError(exc: ImageCaptureException) {
                    Log.e(TAG, "Photo capture failed: ${exc.message}", exc)
                }

                override  fun onCaptureSuccess(image: ImageProxy) {
                    progressBar.visibility = VISIBLE
                    progressBar.progress = 0

                    val imageBitmap = imageProxyToBitmap(image)!!

                    val bos = ByteArrayOutputStream()

                    imageBitmap.compress(Bitmap.CompressFormat.JPEG, 100, bos)

                    val ba = bos.toByteArray()

                    val length = ba.size

                    val bis = ByteArrayInputStream(ba)

                    if (currentState == "image") {
                        Fuel.upload("$serverAddress/image")
                            .add(BlobDataPart(bis, name = "data", filename = "image.jpg", contentLength = length.toLong()))
                            .requestProgress{readBytes, totalBytes ->
                                val progress = readBytes.toFloat() / totalBytes.toFloat() * 50
                                h.post {progressBar.progress = progress.toInt()}
                            }
                            .responseProgress {readBytes, totalBytes ->
                                val progress = readBytes.toFloat() / totalBytes.toFloat() * 50 + 50
                                h.post {progressBar.progress = progress.toInt()}
                            }
                            .response { result ->
                                val (bytes, error) = result
                                if (error != null) {
                                    println("[error] $error")
                                } else {
                                    val res = BitmapFactory.decodeByteArray(bytes, 0, bytes!!.count())
                                    //runOnUiThread {imagePreview.setImageBitmap(res)}
                                    h.post {
                                        imagePreview.setImageBitmap(res)
                                        progressBar.visibility = GONE
                                        if (res.width > res.height) {
                                            imageWidth = 600
                                            imageHeight = res.height * imageWidth / res.width
                                        } else {
                                            imageHeight = 600
                                            imageWidth = res.width * imageHeight / res.height
                                        }
                                        currentState = "view"
                                    }
                                }
                            }
                    }
                    else if (currentState == "view") {
                        Fuel.upload("$serverAddress/view")
                            .add(BlobDataPart(bis, name = "data", filename = "image.jpg", contentLength = length.toLong()))
                            .add(InlineDataPart(imageHeight.toString(), name = "height"))
                            .add(InlineDataPart(imageWidth.toString(), name = "width"))
                            .requestProgress{readBytes, totalBytes ->
                                val progress = readBytes.toFloat() / totalBytes.toFloat() * 50
                                h.post {progressBar.progress = progress.toInt()}
                            }
                            .responseProgress {readBytes, totalBytes ->
                                val progress = readBytes.toFloat() / totalBytes.toFloat() * 50 + 50
                                h.post {progressBar.progress = progress.toInt()}
                            }
                            .response { result ->
                                val (bytes, error) = result
                                if (error != null) {
                                    println("[error] $error")
                                } else {
                                    h.post {
                                        imagePreview.setImageDrawable(null)
                                        progressBar.visibility = GONE
                                        currentState = "image"
                                    }
                                }
                            }
                    }
                    image.close()

                }
            }
        )
    }


    private fun allPermissionsGranted() = REQUIRED_PERMISSIONS.all {
        ContextCompat.checkSelfPermission(baseContext, it) == PackageManager.PERMISSION_GRANTED
    }


    private fun imageProxyToBitmap(image: ImageProxy): Bitmap? {
        val planeProxy = image.planes[0]
        val buffer = planeProxy.buffer
        val bytes = ByteArray(buffer.remaining())
        buffer[bytes]
        return BitmapFactory.decodeByteArray(bytes, 0, bytes.size)
    }

    companion object {
        private const val TAG = "CameraXBasic"
        private const val FILENAME_FORMAT = "yyyy-MM-dd-HH-mm-ss-SSS"
        private const val REQUEST_CODE_PERMISSIONS = 10
        private val REQUIRED_PERMISSIONS = arrayOf(Manifest.permission.CAMERA, Manifest.permission.INTERNET)
    }
}
