package dev.pulse.config;

import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;

@Controller
public class SpaForwardController {

    @RequestMapping(value = {
        "/",
        "/login",
        "/register",
        "/settings",
        "/settings/**",
        "/events",
        "/events/**",
        "/event/**",
        "/poll/**",
        "/me",
        "/me/**",
        "/admin",
        "/admin/**"
    })
    public String forward() {
        return "forward:/index.html";
    }
}