package com.fei.twitterjavaapi.service;

import com.azure.storage.blob.BlobClient;
import com.azure.storage.blob.BlobContainerClient;
import com.azure.storage.blob.BlobServiceClient;
import com.azure.storage.blob.batch.BlobBatch;
import com.azure.storage.blob.batch.BlobBatchClient;
import com.azure.storage.blob.batch.BlobBatchClientBuilder;
import com.azure.storage.blob.models.BlobHttpHeaders;
import com.azure.storage.blob.models.DeleteSnapshotsOptionType;
import com.fei.twitterjavaapi.exception.AppException;
import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class FileStorageService {

    private final BlobServiceClient blobServiceClient;
    private BlobBatchClient blobBatchClient;

    @Value("${spring.cloud.azure.storage.blob.container-name}")
    private String containerName;

    @PostConstruct
    public void init() {
        // Build the Batch Client using the authenticated Service Client
        this.blobBatchClient = new BlobBatchClientBuilder(blobServiceClient).buildClient();
    }

    public String uploadFile(MultipartFile file) {
        if (file == null || file.isEmpty()) return null;

        // 1. Generate Unique Filename
        // e.g., "550e8400-e29b-41d4-a716-446655440000_cat_video.mp4"
        String filename = UUID.randomUUID() + "_" + file.getOriginalFilename();

        // 2. Get Blob Client
        BlobContainerClient containerClient = blobServiceClient.getBlobContainerClient(containerName);
        BlobClient blobClient = containerClient.getBlobClient(filename);

        try {
            // 3. Set Content-Type (Essential for Video Streaming)
            BlobHttpHeaders headers = new BlobHttpHeaders().setContentType(file.getContentType());

            // 4. Upload & Apply Headers
            blobClient.upload(file.getInputStream(), file.getSize(), true);
            blobClient.setHttpHeaders(headers);

            log.info("Uploaded file to Azure: {}", filename);
            return blobClient.getBlobUrl();

        } catch (IOException e) {
            log.error("Failed to upload file to Azure", e);
            throw new AppException("Failed to upload file", HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    public void deleteFile(String fileUrl) {
        if (fileUrl == null || fileUrl.isEmpty()) {
            return;
        }

        try {
            // 1. Extract filename from the full URL
            String filename = fileUrl.substring(fileUrl.lastIndexOf("/") + 1);

            // 2. Get Client
            BlobContainerClient containerClient = blobServiceClient.getBlobContainerClient(containerName);
            BlobClient blobClient = containerClient.getBlobClient(filename);

            // 3. Delete (use deleteIfExists to avoid errors if already gone)
            blobClient.deleteIfExists();
            log.info("Deleted file from Azure: {}", filename);

        } catch (RuntimeException e) {
            // Log it but DO NOT throw exception.
            // Reason: If Azure fails, we still want the Tweet to be deleted from the DB.
            log.error("Failed to delete file from Azure: {}", fileUrl, e);
        }
    }

    // BATCH DELETE
    public void deleteFiles(List<String> rawFileUrls) {
        if (rawFileUrls == null || rawFileUrls.isEmpty()) return;

        // Azure Batch limit is 256 operations per request
        final int BATCH_SIZE = 256;
        List<String> currentBatchFilenames = new ArrayList<>();

        for (String rawUrl : rawFileUrls) {
            try {
                // 1. Extract ONLY the filename.
                String filename = rawUrl.substring(rawUrl.lastIndexOf("/") + 1);
                currentBatchFilenames.add(filename);

                if (currentBatchFilenames.size() == BATCH_SIZE) {
                    processBatchDelete(currentBatchFilenames);
                    currentBatchFilenames.clear();
                }
            } catch (Exception e) {
                log.warn("Skipping malformed URL parsing: {}", rawUrl);
            }
        }

        // Process remaining
        if (!currentBatchFilenames.isEmpty()) {
            processBatchDelete(currentBatchFilenames);
        }
    }

    private void processBatchDelete(List<String> filenames) {
        try {
            // 1. Create a new Batch Object
            BlobBatch batch = blobBatchClient.getBlobBatch();

            // 2. Add "Delete" operations to the batch using Container + Filename
            for (String filename : filenames) {
                batch.deleteBlob(containerName, filename, DeleteSnapshotsOptionType.INCLUDE, null);
            }

            // 3. Submit the Batch (One HTTP Request)
            blobBatchClient.submitBatch(batch);

            log.info("Batch deleted {} files successfully", filenames.size());

        } catch (Exception e) {
            // 4. Fallback Strategy
            log.warn("Batch delete failed (possibly due to missing file). Switching to singular delete fallback.");
            for (String filename : filenames) {
                deleteFileByName(filename);
            }
        }
    }

    // Helper for the fallback
    private void deleteFileByName(String filename) {
        try {
            blobServiceClient.getBlobContainerClient(containerName)
                    .getBlobClient(filename)
                    .deleteIfExists();
        } catch (Exception ex) {
            log.error("Failed to delete blob during fallback: {}", filename, ex);
        }
    }
}