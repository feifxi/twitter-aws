package com.fei.twitterjavaapi.security;

import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpMethod;
import org.springframework.security.config.Customizer;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

@Configuration
@EnableWebSecurity
@RequiredArgsConstructor
public class SecurityConfig {

    private final JwtAuthenticationFilter jwtAuthFilter;
    private final CustomAuthEntryPoint customAuthEntryPoint;
    private final CustomAccessDeniedHandler customAccessDeniedHandler;

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
                // 1. CORS: Uses separate "CorsConfigurationSource" bean automatically
                .cors(Customizer.withDefaults())

                // 2. Disable CSRF (Stateful protection not needed for stateless JWT APIs)
                .csrf(AbstractHttpConfigurer::disable)

                // 3. Define Endpoint Rules
                .authorizeHttpRequests(auth -> auth
                        // SPECIFIC RESTRICTIONS (Must come BEFORE generic wildcards)
                        .requestMatchers(HttpMethod.GET, "/api/v1/auth/me").authenticated()
                        .requestMatchers(HttpMethod.GET, "/api/v1/feeds/following").authenticated()

                        // Public Endpoints
                        .requestMatchers("/error").permitAll()
                        .requestMatchers("/api/v1/auth/**").permitAll()

                        // Public Read Access (GET)
                        // Allow guests to see the feed and single tweets
                        .requestMatchers(HttpMethod.GET, "/api/v1/feeds/**").permitAll()
                        .requestMatchers(HttpMethod.GET, "/api/v1/tweets/**").permitAll()
                        .requestMatchers(HttpMethod.GET, "/api/v1/users/**").permitAll()
                        .requestMatchers(HttpMethod.GET, "/api/v1/search/**").permitAll()
                        .requestMatchers(HttpMethod.GET, "/api/v1/discovery/**").permitAll()

                        // Swagger UI (Optional)
                        .requestMatchers("/swagger-ui/**", "/v3/api-docs/**").permitAll()

                        // Everything else (POST/PUT/DELETE) -> Authenticated
                        .anyRequest().authenticated()
                )

                // 4.  Exception Handling
                .exceptionHandling(e -> e
                        .authenticationEntryPoint(customAuthEntryPoint)     // 401
                        .accessDeniedHandler(customAccessDeniedHandler)     // 403
                )

                // 5. Stateless (No Session/Cookies created by Spring)
                .sessionManagement(sess -> sess.sessionCreationPolicy(SessionCreationPolicy.STATELESS))

                // 6. JWT Filter
                .addFilterBefore(jwtAuthFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }
}